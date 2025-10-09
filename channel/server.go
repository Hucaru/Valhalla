package channel

import (
	"context"
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

type players []*Player

func (p players) getFromConn(conn mnet.Client) (*Player, error) {
	for _, v := range p {
		if v.Conn == conn {
			return v, nil
		}
	}
	return nil, fmt.Errorf("Player not found for connection")
}

func (p players) getFromName(name string) (*Player, error) {
	for _, v := range p {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, fmt.Errorf("Player not found for Name: %s", name)
}

func (p players) getFromID(id int32) (*Player, error) {
	for _, v := range p {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, fmt.Errorf("Player not found for ID: %d", id)
}

// RemoveFromConn removes the Player based on the connection
func (p *players) removeFromConn(conn mnet.Client) error {
	i := -1
	for j, v := range *p {
		if v.Conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("Player not found for removal")
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
	id                byte
	worldName         string
	dispatch          chan func()
	world             mnet.Server
	ip                []byte
	port              int16
	maxPop            int16
	migrating         []mnet.Client
	players           players
	channels          [20]internal.Channel
	cashShop          internal.CashShop
	fields            map[int32]*field
	header            string
	npcChat           map[mnet.Client]*npcChatController
	npcScriptStore    *scriptStore
	eventCtrl         map[string]*eventScriptController
	eventScriptStore  *scriptStore
	portalCtrl        map[mnet.Client]*portalScriptController
	portalScriptStore *scriptStore
	parties           map[int32]*party
	guilds            map[int32]*guild
	rates             rates
}

// Initialise the server
func (server *Server) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase, dropsJson, reactorJson, reactorDropsJson string) {
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

	start := time.Now()
	if err := populateDropTable(dropsJson); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	log.Println("Loaded and parsed drop data in", elapsed)

	start = time.Now()
	if err := populateReactorTable(reactorJson); err != nil {
		log.Fatal(err)
	}
	elapsed = time.Since(start)
	log.Println("Loaded and parsed reactor data in", elapsed)

	start = time.Now()
	if err := populateReactorDropTable(reactorDropsJson); err != nil {
		log.Fatal(err)
	}
	elapsed = time.Since(start)
	log.Println("Loaded and parsed reactor drop data in", elapsed)

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
	server.guilds = make(map[int32]*guild)
}

func (server *Server) loadScripts() {
	server.npcChat = make(map[mnet.Client]*npcChatController)
	server.portalCtrl = make(map[mnet.Client]*portalScriptController)
	server.eventCtrl = make(map[string]*eventScriptController)

	server.npcScriptStore = createScriptStore("scripts/npc", server.dispatch) // make folder a config param
	start := time.Now()
	_ = server.npcScriptStore.loadScripts()
	elapsed := time.Since(start)
	log.Println("Loaded npc scripts in", elapsed)
	go server.npcScriptStore.monitor(func(name string, program *goja.Program) {})

	server.portalScriptStore = createScriptStore("scripts/portals", server.dispatch) // make folder a config param
	start = time.Now()
	_ = server.portalScriptStore.loadScripts()
	elapsed = time.Since(start)
	log.Println("Loaded portal scripts in", elapsed)
	go server.portalScriptStore.monitor(func(name string, program *goja.Program) {})

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
			p.Send(packetHideCountdown())
		} else {
			p.Send(packetShowCountdown(t))
		}
	}
}

// SendLostWorldConnectionMessage - Send message to players alerting them of whatever they do it won't be saved
func (server *Server) SendLostWorldConnectionMessage() {
	for _, p := range server.players {
		p.Send(packetMessageNotice("Cannot connect to world server, any action from the point until the countdown disappears won't be processed"))
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

	defer func() {
		conn.Cleanup()
		common.MetricsGauges["player_count"].With(prometheus.Labels{
			"channel": strconv.Itoa(int(server.id)),
			"world":   server.worldName,
		}).Dec()
	}()

	if field, ok := server.fields[plr.mapID]; ok {
		if inst, ierr := field.getInstance(plr.inst.id); ierr == nil {
			if remErr := inst.removePlayer(plr); remErr != nil {
				log.Println(remErr)
			}
		}
	}

	plr.Logout()

	delete(server.npcChat, conn)

	if remPlrErr := server.players.removeFromConn(conn); remPlrErr != nil {
		log.Println(remPlrErr)
	}

	if idx := func() int {
		for i, v := range server.migrating {
			if v == conn {
				return i
			}
		}
		return -1
	}(); idx > -1 {
		server.migrating = append(server.migrating[:idx], server.migrating[idx+1:]...)
	}

	var guildID int32 = 0
	if plr.guild != nil {
		guildID = plr.guild.id
	}
	server.world.Send(internal.PacketChannelPlayerDisconnect(plr.ID, plr.Name, guildID))

	if _, dbErr := common.DB.Exec("UPDATE characters SET channelID=? WHERE ID=?", -1, plr.ID); dbErr != nil {
		log.Println(dbErr)
	}
	if _, dbErr := common.DB.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID()); dbErr != nil {
		log.Println("Unable to complete logout for ", conn.GetAccountID())
	}
}

func (server *Server) flushPlayers() {
	for _, p := range server.players {
		if p == nil {
			continue
		}
		FlushNow(p)
	}
}

// CheckpointAll now uses the saver to flush debounced/coalesced deltas for every Player.
func (server *Server) CheckpointAll() {
	if server.dispatch == nil {
		return
	}
	done := make(chan struct{})
	server.dispatch <- func() {
		server.flushPlayers()
		close(done)
	}
	<-done
}

// startAutosave periodically flushes deltas via the saver.
func (server *Server) StartAutosave(ctx context.Context) {
	if server.dispatch == nil {
		return
	}
	const interval = 30 * time.Second

	var scheduleNext func()
	scheduleNext = func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.AfterFunc(interval, func() {
			server.dispatch <- func() {
				server.flushPlayers()
				scheduleNext()
			}
		})
	}

	server.dispatch <- func() {
		server.flushPlayers()
		scheduleNext()
	}
}
