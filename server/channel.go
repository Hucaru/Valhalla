package server

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // don't need full import

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

// ChannelServer state
type ChannelServer struct {
	id        byte
	db        *sql.DB
	dispatch  chan func()
	world     mnet.Server
	ip        []byte
	port      int16
	maxPop    int16
	migrating []mnet.Client // TODO: switch to slice
	players   entity.Players
	channels  [20]channel
	fields    map[int32]*entity.Field
	header    string
}

// Initialise the server
func (server *ChannelServer) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

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

	server.fields = make(map[int32]*entity.Field)

	for fieldID, nxMap := range nx.GetMaps() {

		server.fields[fieldID] = &entity.Field{
			ID:      fieldID,
			Data:    nxMap,
			Players: &server.players,
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
	server.id = reader.ReadByte()
	log.Println("Registered as channel", server.id)
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
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()
	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())
	inst.RemovePlayer(player)

	_, err = server.db.Exec("UPDATE characters SET channelID=? WHERE id=?", -1, player.Char().ID())

	if err != nil {
		log.Println(err)
		return
	}

	server.players.RemoveFromConn(conn)

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
}
