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
func (p *players) removeFromConn(conn mnet.Client) error {
	i := -1

	for j, v := range *p {
		if v.conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("Could not find Data")
	}

	(*p)[i] = (*p)[len((*p))-1]
	(*p) = (*p)[:len((*p))-1]

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
	npcChat          map[mnet.Client]*npcScriptController
	npcScriptStore   *scriptStore
	eventCtrl        map[string]*eventScriptController
	eventScriptStore *scriptStore
	parties          map[int32]*party
	rates            rates
}

// Initialise the server
func (server *Server) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

	err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase)

	if err != nil {
		log.Fatal(err)
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

	log.Println("Initialised game state")

	common.MetricsGauges["player_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "player_count",
		Help: "Number of players in this channel",
	}, []string{"channel", "world"})

	prometheus.MustRegister(common.MetricsGauges["player_count"])
	common.StartMetrics()
	log.Println("Started serving metrics on :" + common.MetricsPort)

	server.loadScripts()

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

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	err = inst.removePlayer(plr)

	if err != nil {
		log.Println(err)
	}

	err = plr.save()

	if err != nil {
		log.Println(err)
	}

	if _, ok := server.npcChat[conn]; ok {
		delete(server.npcChat, conn)
	}

	server.players.removeFromConn(conn)

	index := -1

	for i, v := range server.migrating {
		if v == conn {
			index = i
		}
	}

	if index > -1 {
		server.migrating = append(server.migrating[:index], server.migrating[index+1:]...)
	} else {
		server.world.Send(internal.PacketChannelPlayerDisconnect(plr.id, plr.name))

		_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE id=?", -1, plr.id)

		if err != nil {
			log.Println(err)
		}

		_, err := common.DB.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}

	conn.Cleanup()

	common.MetricsGauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Dec()
}

// SetScrollingHeaderMessage that appears at the top of game window
// func (server *Server) SetScrollingHeaderMessage(msg string) {
// 	server.header = msg
// 	for _, v := range server.players {
// 		v.send(message.PacketMessageScrollingHeader(msg))
// 	}
// }
