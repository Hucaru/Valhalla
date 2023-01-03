package channel

import (
	"encoding/binary"
	"fmt"
	"github.com/Hucaru/Valhalla/common/db"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	"github.com/pemistahl/lingua-go"
	"google.golang.org/protobuf/proto"
	"log"
	"strconv"
	"time"

	"github.com/dop251/goja"
	_ "github.com/go-sql-driver/mysql" // don't need full import
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type players map[string]*player

func (p players) getFromConn(conn mnet.Client) (*player, error) {
	//for _, v := range p {
	//	if v.conn == conn {
	//		return v, nil
	//	}
	//}

	plr, ok := p[conn.GetPlayer().UId]
	if ok {
		return plr, nil
	}

	return new(player), fmt.Errorf("Could not retrieve Data")
}

// GetFromName retrieve the Data from the connection
func (p players) getFromName(name string) (*player, error) {
	for _, v := range p {
		if v.name == name {
			return v, nil
		}
	}

	return new(player), fmt.Errorf("Could not retrieve Data")
}

// GetFromID retrieve the Data from the connection
func (p players) getFromID(id int32) (*player, error) {
	for _, v := range p {
		if v.id == id {
			return v, nil
		}
	}

	return new(player), fmt.Errorf("Could not retrieve Data")
}

// RemoveFromConn removes the Data based on the connection
//func (p *players) removeFromConn(conn mnet.Client) error {
//	i := -1
//	for j, v := range *p {
//		if v.conn == conn {
//			i = j
//			break
//		}
//	}
//
//	if i == -1 {
//		return fmt.Errorf("Could not find Data")
//	}
//
//	(*p)[i] = (*p)[len((*p))-1]
//	(*p) = (*p)[:len((*p))-1]
//
//	return nil
//}

type rates struct {
	exp   float32
	drop  float32
	mesos float32
}

// Server state
type Server struct {
	id               byte
	worldName        string
	dispatch         chan func()
	world            mnet.Server
	ip               []byte
	port             int16
	maxPop           int16
	migrating        []mnet.Client
	players          players
	channels         [20]internal.Channel
	fields           map[int32]*field
	header           string
	npcChat          map[mnet.Client]*npcScriptController
	npcScriptStore   *scriptStore
	eventCtrl        map[string]*eventScriptController
	eventScriptStore *scriptStore
	parties          map[int32]*party
	rates            rates
	account          *model.Character
	langDetector     lingua.LanguageDetector
	mapGrid          map[int]map[int][]*player //(y,x)[data]
}

// Initialize the server
func (server *Server) Initialize(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

	err := db.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase)

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Connected to database")

	server.fields = make(map[int32]*field)

	for fieldID, nxMap := range nx.GetMaps() {

		server.fields[fieldID] = &field{
			id:       fieldID,
			Data:     nxMap,
			Dispatch: server.dispatch,
		}
		// For safety, as world will override this
		server.rates = rates{
			exp:   1,
			drop:  1,
			mesos: 1,
		}
		server.fields[fieldID].formatFootholds()
		server.fields[fieldID].calculateFieldLimits()
		server.fields[fieldID].createInstance(&server.rates)
	}

	server.clearSessions()

	columns := (constant.LAND_X1 - constant.LAND_X2) / constant.LAND_VIEW_RANGE
	rows := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	x := make(map[int]map[int][]*player, columns)

	for i := 0; i < columns; i++ {
		y := make(map[int][]*player, rows)

		for j := 0; j < rows; j++ {
			d := []*player{}
			y[j] = d
		}
		x[i] = y
	}

	server.mapGrid = x

	log.Println("Initialised game state")
	common.MetricsGauges["player_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "player_count",
		Help: "Number of players in this channel",
	}, []string{"channel", "world"})

	prometheus.MustRegister(common.MetricsGauges["player_count"])
	common.StartMetrics()
	log.Println("Started serving metrics on :" + common.MetricsPort)

	server.loadScripts()

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages([]lingua.Language{
			lingua.English,
			lingua.Korean,
			lingua.Thai,
		}...).
		Build()

	server.langDetector = detector

	server.parties = make(map[int32]*party)
}

func (server *Server) loadScripts() {
	server.npcChat = make(map[mnet.Client]*npcScriptController)
	server.eventCtrl = make(map[string]*eventScriptController)

	server.npcScriptStore = createScriptStore("scripts/npc", server.dispatch) // make folder a config param
	start := time.Now()
	server.npcScriptStore.loadScripts()
	elapsed := time.Since(start)
	log.Println("Loaded npc scripts in", elapsed)
	go server.npcScriptStore.monitor(func(name string, program *goja.Program) {})

	server.eventScriptStore = createScriptStore("scripts/event", server.dispatch) // make folder a config param
	start = time.Now()
	server.eventScriptStore.loadScripts()
	elapsed = time.Since(start)
	log.Println("Loaded event scripts in", elapsed)

	go server.eventScriptStore.monitor(func(name string, program *goja.Program) {
		if controller, ok := server.eventCtrl[name]; ok && controller != nil {
			controller.Terminate()
		}

		if program == nil {
			if _, ok := server.eventCtrl[name]; ok {
				delete(server.eventCtrl, name)
			}

			return
		}

		controller, start, err := createNewEventScriptController(name, program, server.fields, server.dispatch, server.warpPlayer)

		if err != nil || controller == nil {
			return
		}

		server.eventCtrl[name] = controller

		if start {
			controller.init()
		}

	})

	for name, program := range server.eventScriptStore.scripts {
		controller, start, err := createNewEventScriptController(name, program, server.fields, server.dispatch, server.warpPlayer)

		if err != nil {
			continue
		}

		server.eventCtrl[name] = controller

		if start {
			controller.init()
		}
	}
}

// SendCountdownToPlayers - Send a countdown to players that appears as a clock
func (server Server) SendCountdownToPlayers(time int32) {
	for _, p := range server.players {
		if time == 0 {
			p.send(packetHideCountdown())
		} else {
			p.send(packetShowCountdown(time))
		}
	}
}

// SendLostWorldConnectionMessage - Send message to players alerting them of whatever they do it won't be saved
func (server *Server) SendLostWorldConnectionMessage() {
	for _, p := range server.players {
		p.send(packetMessageNotice("Cannot connect to world server, any action from the point until the countdown disappears won't be processed"))
	}
}

// RegisterWithWorld server
func (server *Server) RegisterWithWorld(conn mnet.Server, ip []byte, port int16, maxPop int16) {
	server.world = conn
	server.ip = ip
	server.port = port
	server.maxPop = maxPop

	server.registerWithWorld()
}

func (server *Server) registerWithWorld() {
	p := mpacket.CreateInternal(opcode.ChannelNew)
	p.WriteBytes(server.ip)
	p.WriteInt16(server.port)
	p.WriteInt16(server.maxPop)
	server.world.Send(p)
}

func (server *Server) clearSessions() {
	err := db.ResetLoginState(false)
	if err != nil {
		log.Println("ERROR LOGOUT PLAYER_ID", err)
	}
}

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	if _, ok := server.npcChat[conn]; ok {
		delete(server.npcChat, conn)
	}

	x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)

	for i := 0; i < len(server.mapGrid[x][y]); i++ {

		if fmt.Sprintf("%p", server.mapGrid[x][y][i]) == fmt.Sprintf("%p", plr) {
			server.mapGrid[x][y][i] = server.mapGrid[x][y][len(server.mapGrid[x][y])-1] // Copy last element to index i.
			server.mapGrid[x][y][len(server.mapGrid[x][y])-1] = nil                     // Erase last element (write zero value).
			server.mapGrid[x][y] = server.mapGrid[x][y][:len(server.mapGrid[x][y])-1]   // Truncate slice.

			break
		}
	}
	server.removePlayer(conn.GetPlayer().UId)

	err1 := db.UpdateLoginState(conn.GetPlayer().UId, false)
	if err1 != nil {
		log.Println("ERROR LOGOUT PLAYER_ID", conn.GetPlayer().UId)
	}

	err2 := db.UpdateMovement(
		conn.GetPlayer().CharacterID,
		conn.GetPlayer().Character.PosX,
		conn.GetPlayer().Character.PosY,
		conn.GetPlayer().Character.PosZ,
		conn.GetPlayer().Character.RotX,
		conn.GetPlayer().Character.RotY,
		conn.GetPlayer().Character.RotZ,
	)

	if err2 != nil {
		log.Println("ERROR UpdateMovement disconnect", err2)
	}

	msg, errR := makeDisconnectedResponse(conn.GetPlayer().UId)
	if errR == nil {
		x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
		loggedPlayers := server.getPlayersOnGrids(x, y, conn.GetPlayer().UId)

		for i := 0; i < len(loggedPlayers); i++ {
			loggedPlayers[i].conn.Send(msg)
		}
	}

	log.Println("Client at", conn, "UID", conn.GetPlayer().UId, "disconnected")

	conn.Cleanup()
	conn = nil
	common.MetricsGauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Dec()
}

func makeDisconnectedResponse(uUID string) ([]byte, error) {
	r := new(mc_metadata.C2P_RequestLogoutUser)
	r.UuId = uUID

	out, err := proto.Marshal(r)
	if err != nil {
		log.Println("Failed to marshal object:", err)
		return nil, err
	}

	result := make([]byte, 0)
	h := make([]byte, 0)
	h = append(h, binary.BigEndian.AppendUint32(h, uint32(len(out)))...)
	h = binary.BigEndian.AppendUint32(h, uint32(constant.P2C_ReportLogoutUser))
	result = append(result, h...)
	result = append(result, out...)

	return result, nil
}

func (server *Server) addPlayer(prl *player) {
	if prl == nil || prl.conn.GetPlayer() == nil {
		return
	}
	if server.players == nil {
		server.players = make(map[string]*player)
	}
	server.players[prl.conn.GetPlayer().UId] = prl
}

func (server *Server) removePlayer(key string) {
	_, ok := server.players[key]
	if ok {
		delete(server.players, key)
	}
}

func (server *Server) getPlayer(key string) *player {
	_, ok := server.players[key]
	if ok {
		return server.players[key]
	}
	return nil
}

// SetScrollingHeaderMessage that appears at the top of game window
// func (server *Server) SetScrollingHeaderMessage(msg string) {
// 	server.header = msg
// 	for _, v := range server.players {
// 		v.send(message.PacketMessageScrollingHeader(msg))
// 	}
// }
