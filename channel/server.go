package channel

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dop251/goja"
	_ "github.com/go-sql-driver/mysql" // don't need full import
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Hucaru/Valhalla/channel/field"
	"github.com/Hucaru/Valhalla/channel/message"
	"github.com/Hucaru/Valhalla/channel/party"
	"github.com/Hucaru/Valhalla/channel/player"
	"github.com/Hucaru/Valhalla/channel/pos"
	"github.com/Hucaru/Valhalla/channel/script"
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type players []*player.Data

func (p players) getFromConn(conn mnet.Client) (*player.Data, error) {
	for _, v := range p {
		if v.Conn() == conn {
			return v, nil
		}
	}

	return new(player.Data), fmt.Errorf("Could not retrieve Data")
}

// GetFromName retrieve the Data from the connection
func (p players) getFromName(name string) (*player.Data, error) {
	for _, v := range p {
		if v.Name() == name {
			return v, nil
		}
	}

	return new(player.Data), fmt.Errorf("Could not retrieve Data")
}

// GetFromID retrieve the Data from the connection
func (p players) getFromID(id int32) (*player.Data, error) {
	for _, v := range p {
		if v.ID() == id {
			return v, nil
		}
	}

	return new(player.Data), fmt.Errorf("Could not retrieve Data")
}

// RemoveFromConn removes the Data based on the connection
func (p *players) removeFromConn(conn mnet.Client) error {
	i := -1

	for j, v := range *p {
		if v.Conn() == conn {
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

// Server state
type Server struct {
	id        byte
	worldName string
	// db               *sql.DB
	dispatch         chan func()
	world            mnet.Server
	ip               []byte
	port             int16
	maxPop           int16
	migrating        []mnet.Client
	players          players
	channels         [20]internal.Channel
	fields           map[int32]*field.Field
	header           string
	npcChat          map[mnet.Client]*script.NpcChatController
	npcScriptStore   *script.Store
	eventCtrl        map[string]*script.EventController
	eventScriptStore *script.Store
	parties          map[int32]*party.Data
}

// Initialise the server
func (server *Server) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

	err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database")

	server.fields = make(map[int32]*field.Field)

	for fieldID, nxMap := range nx.GetMaps() {

		server.fields[fieldID] = &field.Field{
			ID:       fieldID,
			Data:     nxMap,
			Dispatch: server.dispatch,
		}

		server.fields[fieldID].FormatFootholds()
		server.fields[fieldID].CalculateFieldLimits()
		server.fields[fieldID].CreateInstance()
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

	server.parties = make(map[int32]*party.Data)
}

func (server *Server) loadScripts() {
	server.npcChat = make(map[mnet.Client]*script.NpcChatController)
	server.eventCtrl = make(map[string]*script.EventController)

	server.npcScriptStore = script.CreateStore("scripts/npc", server.dispatch) // make folder a config param
	start := time.Now()
	server.npcScriptStore.LoadScripts()
	elapsed := time.Since(start)
	log.Println("Loaded npc scripts in", elapsed)
	go server.npcScriptStore.Monitor(func(name string, program *goja.Program) {})

	server.eventScriptStore = script.CreateStore("scripts/event", server.dispatch) // make folder a config param
	start = time.Now()
	server.eventScriptStore.LoadScripts()
	elapsed = time.Since(start)
	log.Println("Loaded event scripts in", elapsed)

	go server.eventScriptStore.Monitor(func(name string, program *goja.Program) {
		if controller, ok := server.eventCtrl[name]; ok && controller != nil {
			controller.Terminate()
		}

		if program == nil {
			if _, ok := server.eventCtrl[name]; ok {
				delete(server.eventCtrl, name)
			}

			return
		}

		controller, start, err := script.CreateNewEventController(name, program, server.fields, server.dispatch, server.warpPlayer)

		if err != nil || controller == nil {
			return
		}

		server.eventCtrl[name] = controller

		if start {
			controller.Init()
		}

	})

	for name, program := range server.eventScriptStore.Scripts() {
		controller, start, err := script.CreateNewEventController(name, program, server.fields, server.dispatch, server.warpPlayer)

		if err != nil {
			continue
		}

		server.eventCtrl[name] = controller

		if start {
			controller.Init()
		}
	}
}

// SendCountdownToPlayers - Send a countdown to players that appears as a clock
func (server Server) SendCountdownToPlayers(time int32) {
	for _, p := range server.players {
		if time == 0 {
			p.Send(message.PacketHideCountdown())
		} else {
			p.Send(message.PacketShowCountdown(time))
		}
	}
}

// SendLostWorldConnectionMessage - Send message to players alerting them of whatever they do it won't be saved
func (server *Server) SendLostWorldConnectionMessage() {
	for _, p := range server.players {
		p.Send(message.PacketMessageNotice("Cannot connect to world server, any action from the point until the countdown disappears won't be processed"))
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

// HandleServerPacket from world
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelBad:
		server.handleNewChannelBad(conn, reader)
	case opcode.ChannelOk:
		server.handleNewChannelOK(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	case opcode.ChannePlayerConnect:
		server.handlePlayerConnectedNotifications(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnectNotifications(conn, reader)
	case opcode.ChannelPlayerChatEvent:
		server.handleChatEvent(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		server.handleBuddyEvent(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handleNewChannelBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by world server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithWorld()
}

func (server *Server) handleNewChannelOK(conn mnet.Server, reader mpacket.Reader) {
	server.worldName = reader.ReadString(reader.ReadInt16())
	server.id = reader.ReadByte()
	log.Println("Registered as channel", server.id, "on world", server.worldName)

	for _, p := range server.players {
		p.Send(message.PacketMessageNotice("Re-connected to world server as channel " + strconv.Itoa(int(server.id+1))))
		// TODO send largest party id for world server to compare
	}

	accountIDs, err := common.DB.Query("SELECT accountID from characters where channelID = ? and migrationID = -1", server.id)

	if err != nil {
		log.Println(err)
		return
	}

	for accountIDs.Next() {
		var accountID int
		err := accountIDs.Scan(&accountID)

		if err != nil {
			continue
		}

		_, err = common.DB.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", 0, accountID)

		if err != nil {
			log.Println(err)
			return
		}
	}

	accountIDs.Close()

	_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE channelID=?", -1, server.id)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Loged out any accounts still connected to this channel")
}

func (server *Server) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handlePlayerConnectedNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	channelID := reader.ReadByte()
	changeChannel := reader.ReadBool()

	plr, _ := server.players.getFromID(playerID)

	for _, party := range server.parties {
		party.SetPlayerChannel(plr, playerID, false, false, int32(channelID))
	}

	for i, v := range server.players {
		if v.ID() == playerID {
			continue
		} else if v.HasBuddy(playerID) {
			if changeChannel {
				server.players[i].Send(message.PacketBuddyChangeChannel(playerID, int32(channelID)))
				server.players[i].AddOnlineBuddy(playerID, name, int32(channelID))
			} else {
				// send online message card, then update buddy list
				server.players[i].Send(message.PacketBuddyOnlineStatus(playerID, int32(channelID)))
				server.players[i].AddOnlineBuddy(playerID, name, int32(channelID))
			}
		}
	}
}

func (server *Server) handlePlayerDisconnectNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())

	for _, party := range server.parties {
		party.SetPlayerChannel(new(player.Data), playerID, false, true, 0)
	}

	for i, v := range server.players {
		if v.ID() == playerID {
			continue
		} else if v.HasBuddy(playerID) {
			server.players[i].AddOfflineBuddy(playerID, name)
		}
	}
}

func (server *Server) handleBuddyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.Send(message.PacketBuddyReceiveRequest(fromID, fromName, int32(channelID)))
	case 2:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.AddOfflineBuddy(fromID, fromName)
		plr.Send(message.PacketBuddyOnlineStatus(fromID, int32(channelID)))
		plr.AddOnlineBuddy(fromID, fromName, int32(channelID))
	case 3:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.RemoveBuddy(fromID)
	default:
		log.Println("Unknown buddy event type:", op)
	}
}

func (server Server) handleChatEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0: // whispher
		recepientName := reader.ReadString(reader.ReadInt16())
		fromName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		plr, err := server.players.getFromName(recepientName)

		if err != nil {
			return
		}

		plr.Send(message.PacketMessageWhisper(fromName, msg, channelID))

	case 1: // buddy
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(0, fromName, msg))
		}
	case 2: // party
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(1, fromName, msg))
		}
	case 3: // guild
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(2, fromName, msg))
		}
	default:
		log.Println("Unknown chat event type:", op)
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0:
		log.Println("Channel server should not receive party event message type: 0")
	case 1: // new party created
		channelID := reader.ReadByte()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		// TODO: Mystic door information needs to be sent here if the leader has an active door

		newParty := party.NewParty(partyID, plr, channelID, playerID, mapID, job, level, name, int32(server.id))

		server.parties[partyID] = &newParty

		if plr != nil {
			plr.SetParty(&newParty)
			plr.Send(message.PacketPartyCreate(1, -1, -1, pos.New(0, 0, 0)))
		}
	case 2: // leave party
		destroy := reader.ReadBool()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.RemovePlayer(plr, playerID, false)

			if destroy {
				delete(server.parties, partyID)
			}
		}
	case 3: // accept
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		channelID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.AddPlayer(plr, channelID, playerID, name, mapID, job, level)
		}
	case 4: // expel
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.RemovePlayer(plr, playerID, true)
		}
	case 5:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		reader.ReadString(reader.ReadInt16()) // name
		if party, ok := server.parties[partyID]; ok {
			party.UpdateJobLevel(playerID, job, level)
		}
	default:
		log.Println("Unkown party event type:", op)
	}
}

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())
	err = inst.RemovePlayer(plr)

	if err != nil {
		log.Println(err)
	}

	err = plr.Save()

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
		server.world.Send(internal.PacketChannelPlayerDisconnect(plr.ID(), plr.Name()))

		_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE id=?", -1, plr.ID())

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
func (server *Server) SetScrollingHeaderMessage(msg string) {
	server.header = msg
	for _, v := range server.players {
		v.Send(message.PacketMessageScrollingHeader(msg))
	}
}
