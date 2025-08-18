package channel

import (
	"fmt"
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

type players []*player

func (p players) getFromConn(conn mnet.Client) (*player, error) {
	for _, v := range p {
		if v.conn == conn {
			return v, nil
		}
	}
	return nil, fmt.Errorf("player not found for connection")
}

func (p players) getFromName(name string) (*player, error) {
	for _, v := range p {
		if v.name == name {
			return v, nil
		}
	}
	return nil, fmt.Errorf("player not found for name: %s", name)
}

func (p players) getFromID(id int32) (*player, error) {
	for _, v := range p {
		if v.id == id {
			return v, nil
		}
	}
	return nil, fmt.Errorf("player not found for id: %d", id)
}

// RemoveFromConn removes the player based on the connection
func (p *players) removeFromConn(conn mnet.Client) error {
	i := -1
	for j, v := range *p {
		if v.conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("player not found for removal")
	}

	(*p)[i] = (*p)[len(*p)-1]
	*p = (*p)[:len(*p)-1]
	return nil
}

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
	npcChat          map[mnet.Client]*npcChatController
	npcScriptStore   *scriptStore
	eventCtrl        map[string]*eventScriptController
	eventScriptStore *scriptStore
	parties          map[int32]*party
	rates            rates
}

// Initialise the server
func (server *Server) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

	if err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")

	server.fields = make(map[int32]*field)

	// Default rates (world may override later)
	server.rates = rates{
		exp:   1,
		drop:  1,
		mesos: 1,
	}

	for fieldID, nxMap := range nx.GetMaps() {
		server.fields[fieldID] = &field{
			id:       fieldID,
			Data:     nxMap,
			Dispatch: server.dispatch,
		}
		server.fields[fieldID].formatFootholds()
		server.fields[fieldID].calculateFieldLimits()
		server.fields[fieldID].createInstance(&server.rates)
	}

	log.Println("Initialised game state")

	// Register metrics gauge only once per process
	if _, ok := common.MetricsGauges["player_count"]; !ok {
		common.MetricsGauges["player_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "player_count",
			Help: "Number of players in this channel",
		}, []string{"channel", "world"})
		prometheus.MustRegister(common.MetricsGauges["player_count"])
	}

	common.StartMetrics()
	log.Println("Started serving metrics on :" + common.MetricsPort)

	server.loadScripts()

	server.parties = make(map[int32]*party)
}

func (server *Server) loadScripts() {
	server.npcChat = make(map[mnet.Client]*npcChatController)
	server.eventCtrl = make(map[string]*eventScriptController)

	server.npcScriptStore = createScriptStore("scripts/npc", server.dispatch) // make folder a config param
	start := time.Now()
	_ = server.npcScriptStore.loadScripts()
	elapsed := time.Since(start)
	log.Println("Loaded npc scripts in", elapsed)
	go server.npcScriptStore.monitor(func(name string, program *goja.Program) {})

	server.eventScriptStore = createScriptStore("scripts/event", server.dispatch) // make folder a config param
	start = time.Now()
	_ = server.eventScriptStore.loadScripts()
	elapsed = time.Since(start)
	log.Println("Loaded event scripts in", elapsed)

	go server.eventScriptStore.monitor(func(name string, program *goja.Program) {
		if controller, ok := server.eventCtrl[name]; ok && controller != nil {
			controller.Terminate()
		}

		if program == nil {
			delete(server.eventCtrl, name)
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
func (server *Server) SendCountdownToPlayers(t int32) {
	for _, p := range server.players {
		if t == 0 {
			p.send(packetHideCountdown())
		} else {
			p.send(packetShowCountdown(t))
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

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	// Always perform connection cleanup and metrics decrement
	defer func() {
		conn.Cleanup()
		common.MetricsGauges["player_count"].With(prometheus.Labels{
			"channel": strconv.Itoa(int(server.id)),
			"world":   server.worldName,
		}).Dec()
	}()

	// Try to remove the player from their current instance
	if field, ok := server.fields[plr.mapID]; ok {
		if inst, ierr := field.getInstance(plr.inst.id); ierr == nil {
			if remErr := inst.removePlayer(plr); remErr != nil {
				log.Println(remErr)
			}
		}
	}

	// Persist character state
	if saveErr := plr.save(); saveErr != nil {
		log.Println(saveErr)
	}

	// Tear down any active NPC chat state
	delete(server.npcChat, conn)

	// Remove from in-memory player list
	if remPlrErr := server.players.removeFromConn(conn); remPlrErr != nil {
		log.Println(remPlrErr)
	}

	// Remove from migrating slice if present
	migratingIdx := -1
	for i, v := range server.migrating {
		if v == conn {
			migratingIdx = i
			break
		}
	}

	if migratingIdx > -1 {
		server.migrating = append(server.migrating[:migratingIdx], server.migrating[migratingIdx+1:]...)
		return
	}

	// Not migrating: notify world and clear DB session state
	server.world.Send(internal.PacketChannelPlayerDisconnect(plr.id, plr.name))

	if _, dbErr := common.DB.Exec("UPDATE characters SET channelID=? WHERE id=?", -1, plr.id); dbErr != nil {
		log.Println(dbErr)
	}

	if _, dbErr := common.DB.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID()); dbErr != nil {
		log.Println("Unable to complete logout for ", conn.GetAccountID())
	}
}

// SetScrollingHeaderMessage that appears at the top of game window
// func (server *Server) SetScrollingHeaderMessage(msg string) {
// 	server.header = msg
// 	for _, v := range server.players {
// 		v.send(message.PacketMessageScrollingHeader(msg))
// 	}
// }
