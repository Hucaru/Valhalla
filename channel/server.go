package channel

import (
	"encoding/binary"
	"fmt"
	"github.com/Hucaru/Valhalla/common/db"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/common/manager"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	_ "github.com/go-sql-driver/mysql" // don't need full import
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pemistahl/lingua-go"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
	"log"
	"runtime"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type players map[string]*player

var gorotinesManager = manager.Init()

func (p players) getFromConn(conn *mnet.Client) (*player, error) {
	plr, ok := p[conn.GetPlayer().UId]
	if ok {
		return plr, nil
	}

	return new(player), fmt.Errorf("Could not retrieve Data")
}

func (p players) getFromConnByUID(uID string) (*player, error) {
	plr, ok := p[uID]
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

type PlayerMovement struct {
	name string
	x    float32
	y    float32
}

type RequestedParam struct {
	Num    uint32
	Reader mpacket.Reader
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
	npcChat          map[string]*npcScriptController
	npcScriptStore   *scriptStore
	eventCtrl        map[string]*eventScriptController
	eventScriptStore *scriptStore
	parties          map[int32]*party
	rates            rates
	account          *model.Character
	langDetector     lingua.LanguageDetector
	mapGrid          [][]map[int]*player //(y,x)[data]
	fMovePlayers     []PlayerMovement    //(y,x)[data]

	gridMgr       manager.GridManager
	clients       cmap.ConcurrentMap[string, *mnet.Client]
	playerActions cmap.ConcurrentMap[string, chan RequestedParam]

	// Kioni
	PlayerActionHandler map[uint32]func(*mnet.Client, mpacket.Reader)
}

// Initialize the server
func (server *Server) Initialize(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {

	// Kioni
	server.PlayerActionHandler = make(map[uint32]func(*mnet.Client, mpacket.Reader), 0)

	server.PlayerActionHandler[constant.C2P_RequestLoginUser] = server.playerConnect
	server.PlayerActionHandler[constant.C2P_RequestMoveStart] = server.playerMovementStart
	server.PlayerActionHandler[constant.C2P_RequestMove] = server.playerMovement
	server.PlayerActionHandler[constant.C2P_RequestMoveEnd] = server.playerMovementEnd
	server.PlayerActionHandler[constant.C2P_RequestLogoutUser] = server.playerLogout
	server.PlayerActionHandler[constant.C2P_RequestPlayerInfo] = server.playerInfo
	server.PlayerActionHandler[constant.C2P_RequestAllChat] = server.chatSendAll
	server.PlayerActionHandler[constant.C2P_RequestWhisper] = server.chatSendWhisper
	server.PlayerActionHandler[constant.C2P_RequestRegionChat] = server.chatSendRegion
	server.PlayerActionHandler[constant.OnDisconnected] = server.ClientDisconnected

	runtime.GOMAXPROCS(runtime.NumCPU())
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

	columns := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	rows := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	server.fMovePlayers = []PlayerMovement{}

	x := make([][]map[int]*player, columns)

	for i := 0; i < columns; i++ {
		y := make([]map[int]*player, rows)

		for j := 0; j < rows; j++ {
			d := make(map[int]*player)
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

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages([]lingua.Language{
			lingua.English,
			lingua.Korean,
			lingua.Thai,
		}...).
		Build()

	server.langDetector = detector

	server.parties = make(map[int32]*party)

	server.gridMgr = manager.GridManager{}
	server.gridMgr.Init()

	server.clients = cmap.New[*mnet.Client]()
	server.playerActions = cmap.New[chan RequestedParam]()
}

func (server *Server) addToEmulateMoving(uid string, plrs []*player) {
	//arr := plrs
	//for i := 0; i < len(arr); i++ {
	//	if plrs[i].conn.GetPlayer().IsBot != 1 {
	//		return
	//	}
	//	server.moveEmulate(arr[i])
	//}
	//arr = nil
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

// Kioni
// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn *mnet.Client, reader mpacket.Reader) {
	//_, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return
	//}
	//
	//if _, ok := server.npcChat[conn]; ok {
	//	delete(server.npcChat, conn)
	//}

	server.removePlayer(conn)

	fmt.Println("NumGoroutine COUNT", runtime.NumGoroutine())
	err1 := db.UpdateLoginState(conn.GetPlayer().UId, false)
	if err1 != nil {
		log.Println("ERROR LOGOUT PLAYER_ID", conn.GetPlayer().UId)
	}

	if conn.GetPlayer().IsBot != 1 {
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
	}

	msg, errR := makeDisconnectedResponse(conn.GetPlayer().UId)
	if errR == nil {
		x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
		loggedPlayers := server.getPlayersOnGrids(x, y, conn.GetPlayer().UId)

		for _, v := range loggedPlayers {
			(*v).Send(msg)
		}

		/*for i := 0; i < len(loggedPlayers); i++ {
			loggedPlayers[i].conn.Send(msg)
		}*/
	}

	//log.Println("Client at", conn, "UID", conn.GetPlayer().UId, "disconnected")

	conn.Cleanup()
	conn = nil
	//common.MetricsGauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Dec()
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

func (server *Server) addPlayer(plr *player) {
	//if plr == nil || plr.conn.GetPlayer() == nil {
	//	return
	//}
	//if server.players == nil {
	//	server.players = make(map[string]*player)
	//}
	//server.addPlayerToGrid(plr, plr.conn.GetPlayer().Character.PosX, plr.conn.GetPlayer().Character.PosY)
	//SomeMapMutex.Lock()
	//server.players[plr.conn.GetPlayer().UId] = plr
	//SomeMapMutex.Unlock()

}

func (server *Server) addPlayerToGrid(plr *player, x1, y1 float32) {
	//if plr == nil {
	//	return
	//}
	//x, y := common.FindGrid(x1, y1)
	//
	//is := false
	//for i := 0; i < len(server.mapGrid[x][y]); i++ {
	//	SomeMapMutex.RLock()
	//	if server.mapGrid[x][y][i] != nil && server.mapGrid[x][y][i].conn.GetPlayer().UId == plr.conn.GetPlayer().UId {
	//		is = true
	//	}
	//	SomeMapMutex.RUnlock()
	//}
	//if !is {
	//	SomeMapMutex.Lock()
	//	server.mapGrid[x][y][len(server.mapGrid[x][y])] = plr
	//	SomeMapMutex.Unlock()
	//}
}

func (server *Server) removePlayer(conn *mnet.Client) {

	server.gridMgr.Remove(conn.GetPlayer().UId)
	server.clients.Remove(conn.GetPlayer().UId)
	//for i := 0; i < len(server.mapGrid); i++ {
	//	x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
	//	SomeMapMutex.RLock()
	//	_, ok := server.mapGrid[x][y][i]
	//	SomeMapMutex.RUnlock()
	//	if ok {
	//		if server.mapGrid[x][y][i].conn.GetPlayer().UId == conn.GetPlayer().UId {
	//			SomeMapMutex.Lock()
	//			delete(server.mapGrid[x][y], i)
	//			SomeMapMutex.Unlock()
	//			break
	//		}
	//	}
	//}
	//
	//SomeMapMutex.RLock()
	//_, ok := server.players[conn.GetPlayer().UId]
	//SomeMapMutex.RUnlock()
	//
	//if ok {
	//	SomeMapMutex.Lock()
	//	delete(server.players, conn.GetPlayer().UId)
	//	SomeMapMutex.Unlock()
	//}

	//server.removeFromMovingLoop(uid)
	//go gorotinesManager.ClearAll()
}

func (server *Server) removePlayerFromGrid(plr map[int]*player, uID string, x1, y1 float32) {
	//if plr == nil {
	//	return
	//}
	//x, y := common.FindGrid(x1, y1)
	//for i := 0; i < len(plr); i++ {
	//
	//	if plr[i] != nil && plr[i].conn.GetPlayer().UId == uID {
	//		delete(server.mapGrid[x][y], i)
	//	}
	//}
}

func (server *Server) removeFromMovingLoop(uid string) {
	for i := 0; i < len(server.fMovePlayers); i++ {
		if server.fMovePlayers[i].name == uid {
			server.fMovePlayers = append(server.fMovePlayers[:i], server.fMovePlayers[i+1:]...)
			//server.removePlayersFromMovementLoop(v[i])
		}
	}

}

func (server *Server) removeFromEmulateMoving(uid string, plrs []*player) {
	for i := 0; i < len(plrs); i++ {
		server.removePlayersFromMovementLoop(plrs[i].conn.GetPlayer().UId)
	}
}

func (server *Server) removePlayersFromMovementLoop(uID string) {
	gorotinesManager.Remove(uID)
}

// SetScrollingHeaderMessage that appears at the top of game window
// func (server *Server) SetScrollingHeaderMessage(msg string) {
// 	server.header = msg
// 	for _, v := range server.players {
// 		v.send(message.PacketMessageScrollingHeader(msg))
// 	}
// }
