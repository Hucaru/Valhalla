package channel

import (
	"context"
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
	players          Players
	channels         [20]internal.Channel
	cashShop         internal.CashShop
	fields           map[int32]*field
	header           string
	npcChat          map[mnet.Client]*npcChatController
	npcScriptStore   *scriptStore
	eventScriptStore *scriptStore
	parties          map[int32]*party
	guilds           map[int32]*guild
	events           map[int32]*event
	rates            rates
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
		server.fields[fieldID].createInstance(&server.rates, server)
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

	if _, ok := common.MetricsCounters["monster_kills_total"]; !ok {
		common.MetricsCounters["monster_kills_total"] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "monster_kills_total",
			Help: "Total number of monsters killed",
		}, []string{"channel", "world", "character_id"})
		prometheus.MustRegister(common.MetricsCounters["monster_kills_total"])
	}

	if _, ok := common.MetricsGauges["ongoing_trades"]; !ok {
		common.MetricsGauges["ongoing_trades"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ongoing_trades",
			Help: "Number of ongoing trades",
		}, []string{"channel", "world"})
		prometheus.MustRegister(common.MetricsGauges["ongoing_trades"])
	}

	if _, ok := common.MetricsGauges["ongoing_minigames"]; !ok {
		common.MetricsGauges["ongoing_minigames"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ongoing_minigames",
			Help: "Number of ongoing minigames",
		}, []string{"channel", "world"})
		prometheus.MustRegister(common.MetricsGauges["ongoing_minigames"])
	}

	if _, ok := common.MetricsGauges["ongoing_npc_interactions"]; !ok {
		common.MetricsGauges["ongoing_npc_interactions"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ongoing_npc_interactions",
			Help: "Number of ongoing NPC script interactions",
		}, []string{"channel", "world"})
		prometheus.MustRegister(common.MetricsGauges["ongoing_npc_interactions"])
	}

	if _, ok := common.MetricsGauges["party_count"]; !ok {
		common.MetricsGauges["party_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "party_count",
			Help: "Number of active parties",
		}, []string{"channel", "world"})
		prometheus.MustRegister(common.MetricsGauges["party_count"])
	}

	common.StartMetrics()
	log.Println("Started serving metrics on :" + common.MetricsPort)

	server.loadScripts()

	server.players = NewPlayers()
	server.parties = make(map[int32]*party)
	server.guilds = make(map[int32]*guild)
	server.events = make(map[int32]*event)

	go scheduleBoats(server)
}

func (server *Server) loadScripts() {
	server.npcChat = make(map[mnet.Client]*npcChatController)
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

	go server.eventScriptStore.monitor(func(name string, program *goja.Program) {})
}

// SendCountdownToPlayers - Send a countdown to players that appears as a clock
func (server *Server) SendCountdownToPlayers(t int32) {
	if t == 0 {
		server.players.broadcast(packetHideCountdown())
	} else {
		server.players.broadcast(packetShowCountdown(t))
	}
}

// SendLostWorldConnectionMessage - Send message to players alerting them of whatever they do it won't be saved
func (server *Server) SendLostWorldConnectionMessage() {
	server.players.broadcast(packetMessageNotice("Cannot connect to world server, any action from the point until the countdown disappears won't be processed"))
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
	plr, err := server.players.GetFromConn(conn)
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

	if plr != nil && (plr.doorMapID != 0 || plr.townDoorMapID != 0) {
		removeMysticDoor(plr)
	}

	if field, ok := server.fields[plr.mapID]; ok {
		if inst, ierr := field.getInstance(plr.inst.id); ierr == nil {
			if remErr := inst.removePlayer(plr, true); remErr != nil {
				log.Println(remErr)
			}
		}
	}

	plr.Logout()

	if _, ok := server.npcChat[conn]; ok {
		delete(server.npcChat, conn)
		server.updateNPCInteractionMetric(-1)
	}

	if remPlrErr := server.players.RemoveFromConn(conn); remPlrErr != nil {
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

// CheckpointAll now uses the saver to flush debounced/coalesced deltas for every Player.
func (server *Server) CheckpointAll(ctx context.Context) {
	if server.dispatch == nil {
		return
	}

	done := make(chan struct{})

	select {
	case <-ctx.Done():
		log.Println("CheckpointAll: cancelled before scheduling flush:", ctx.Err())
		return

	case server.dispatch <- func() {
		server.players.Flush()
		close(done)
	}:
	}

	select {
	case <-ctx.Done():
		log.Println("CheckpointAll: cancelled while waiting for flush:", ctx.Err())
		return
	case <-time.After(10 * time.Second):
		log.Println("CheckpointAll: timed out waiting for dispatcher flush (dispatcher may be stopped)")
		return
	case <-done:
		return
	}
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
			select {
			case <-ctx.Done():
				return
			case server.dispatch <- func() {
				server.players.Flush()
				scheduleNext()
			}:
			}
		})
	}

	select {
	case <-ctx.Done():
		return
	case server.dispatch <- func() {
		server.players.Flush()
		scheduleNext()
	}:
	}
}

func (server *Server) updatePartyMetric() {
	if server.worldName == "" {
		return
	}
	common.MetricsGauges["party_count"].With(prometheus.Labels{
		"channel": strconv.Itoa(int(server.id)),
		"world":   server.worldName,
	}).Set(float64(len(server.parties)))
}

func (server *Server) updateMobKillMetric(charID int32) {
	if server.worldName == "" {
		return
	}
	common.MetricsCounters["monster_kills_total"].With(prometheus.Labels{
		"channel":      strconv.Itoa(int(server.id)),
		"world":        server.worldName,
		"character_id": strconv.Itoa(int(charID)),
	}).Inc()
}

func (server *Server) updateTradeMetric(delta int) {
	if server.worldName == "" {
		return
	}
	common.MetricsGauges["ongoing_trades"].With(prometheus.Labels{
		"channel": strconv.Itoa(int(server.id)),
		"world":   server.worldName,
	}).Add(float64(delta))
}

func (server *Server) updateMinigameMetric(delta int) {
	if server.worldName == "" {
		return
	}
	common.MetricsGauges["ongoing_minigames"].With(prometheus.Labels{
		"channel": strconv.Itoa(int(server.id)),
		"world":   server.worldName,
	}).Add(float64(delta))
}

func (server *Server) updateNPCInteractionMetric(delta int) {
	if server.worldName == "" {
		return
	}
	common.MetricsGauges["ongoing_npc_interactions"].With(prometheus.Labels{
		"channel": strconv.Itoa(int(server.id)),
		"world":   server.worldName,
	}).Add(float64(delta))
}
