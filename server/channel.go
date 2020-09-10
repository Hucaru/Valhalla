package server

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql" // don't need full import
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/message"
	"github.com/Hucaru/Valhalla/server/metrics"
	"github.com/Hucaru/Valhalla/server/player"
	"github.com/Hucaru/Valhalla/server/script"
)

type players []*player.Data

func (p players) getFromConn(conn mnet.Client) (*player.Data, error) {
	for _, v := range p {
		if v.Conn() == conn {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve Data")
}

// GetFromName retrieve the Data from the connection
func (p players) getFromName(name string) (*player.Data, error) {
	for _, v := range p {
		if v.Name() == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve Data")
}

// GetFromID retrieve the Data from the connection
func (p players) getFromID(id int32) (*player.Data, error) {
	for _, v := range p {
		if v.ID() == id {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve Data")
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

// ChannelServer state
type ChannelServer struct {
	id                byte
	worldName         string
	db                *sql.DB
	dispatch          chan func()
	world             mnet.Server
	ip                []byte
	port              int16
	maxPop            int16
	migrating         []mnet.Client
	players           players
	channels          [20]channel
	fields            map[int32]*field.Field
	header            string
	npcChat           map[mnet.Client]*script.NpcChatController
	npcScriptStore    *script.Store
	systemCntrl       map[string]*script.SystemController
	systemScriptStore *script.Store
}

// Initialise the server
func (server *ChannelServer) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

	server.npcChat = make(map[mnet.Client]*script.NpcChatController)
	server.systemCntrl = make(map[string]*script.SystemController)

	server.npcScriptStore = script.CreateStore("scripts/npc", server.dispatch) // make folder a config param
	start := time.Now()
	server.npcScriptStore.LoadScripts()
	elapsed := time.Since(start)
	log.Println("Loaded npc scripts in", elapsed)
	go server.npcScriptStore.Monitor()

	server.systemScriptStore = script.CreateStore("scripts/system", server.dispatch) // make folder a config param
	start = time.Now()
	server.systemScriptStore.LoadScripts()
	elapsed = time.Since(start)
	log.Println("Loaded system scripts in", elapsed)
	// go server.systemScriptStore.Monitor() // system controller takes script as a 1 time event and cannot be hotloaded

	var err error
	server.db, err = sql.Open("mysql", dbuser+":"+dbpassword+"@tcp("+dbaddress+":"+dbport+")/"+dbdatabase)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Connected to database")

	server.fields = make(map[int32]*field.Field)

	for fieldID, nxMap := range nx.GetMaps() {

		server.fields[fieldID] = &field.Field{
			ID:       fieldID,
			Data:     nxMap,
			Dispatch: server.dispatch,
		}

		server.fields[fieldID].CalculateFieldLimits()
		server.fields[fieldID].CreateInstance()
	}

	log.Println("Initialised game state")

	accountIDs, err := server.db.Query("SELECT accountID from characters where channelID = ?", server.id)

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

		_, err = server.db.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", 0, accountID)

		if err != nil {
			log.Println(err)
			return
		}
	}

	accountIDs.Close()

	_, err = server.db.Exec("UPDATE characters SET channelID=? WHERE channelID=?", -1, server.id)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Loged out any accounts still connected to this channel")

	metrics.Gauges["player_count"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "player_count",
		Help: "Number of players in this channel",
	}, []string{"channel", "world"})

	prometheus.MustRegister(metrics.Gauges["player_count"])
	metrics.StartMetrics()
	log.Println("Started serving metrics on :" + metrics.Port)

	for name, program := range server.systemScriptStore.Scripts() {
		controller, err := script.CreateNewSystemController(name, program, server.fields, server.dispatch)

		if err != nil {
			continue
		}

		server.systemCntrl[name] = controller
		controller.Start()
	}
}

// SendCountdownToPlayers - Send a countdown to players that appears as a clock
func (server ChannelServer) SendCountdownToPlayers(time int32) {
	for _, p := range server.players {
		if time == 0 {
			p.Send(message.PacketHideCountdown())
		} else {
			p.Send(message.PacketShowCountdown(time))
		}
	}
}

// SendLostWorldConnectionMessage - Send message to players alerting them of whatever they do it won't be saved
func (server *ChannelServer) SendLostWorldConnectionMessage() {
	for _, p := range server.players {
		p.Send(message.PacketMessageNotice("Cannot connect to world server, any action from the point until the countdown disappears won't be processed"))
	}
}

// RegisterWithWorld server
func (server *ChannelServer) RegisterWithWorld(conn mnet.Server, ip []byte, port int16, maxPop int16) {
	server.world = conn
	server.ip = ip
	server.port = port
	server.maxPop = maxPop

	server.registerWithWorld()
}

func (server *ChannelServer) registerWithWorld() {
	p := mpacket.CreateInternal(opcode.ChannelNew)
	p.WriteBytes(server.ip)
	p.WriteInt16(server.port)
	p.WriteInt16(server.maxPop)
	server.world.Send(p)
}

// HandleServerPacket from world
func (server *ChannelServer) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelBad:
		server.handleNewChannelBad(conn, reader)
	case opcode.ChannelOk:
		server.handleNewChannelOK(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *ChannelServer) handleNewChannelBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by world server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithWorld()
}

func (server *ChannelServer) handleNewChannelOK(conn mnet.Server, reader mpacket.Reader) {
	server.worldName = reader.ReadString(reader.ReadInt16())
	server.id = reader.ReadByte()
	log.Println("Registered as channel", server.id, "on world", server.worldName)
}

func (server *ChannelServer) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].ip = reader.ReadBytes(4)
		server.channels[i].port = reader.ReadInt16()
	}
}

// ClientDisconnected from server
func (server *ChannelServer) ClientDisconnected(conn mnet.Client) {
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

	err = plr.Save(server.db)

	if err != nil {
		log.Println(err)
	}

	_, err = server.db.Exec("UPDATE characters SET channelID=? WHERE id=?", -1, plr.ID())

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
		_, err := server.db.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}

	conn.Cleanup()

	metrics.Gauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Dec()
}

// SetScrollingHeaderMessage that appears at the top of game window
func (server *ChannelServer) SetScrollingHeaderMessage(msg string) {
	server.header = msg
	for _, v := range server.players {
		v.Send(message.PacketMessageScrollingHeader(msg))
	}
}
