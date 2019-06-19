package game

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // don't need full import

	"github.com/Hucaru/Valhalla/constant/opcode"
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
	migrating map[mnet.Client]byte // TODO: switch to slice
	players   map[mnet.Client]*player
	channels  [20]channel
	fields    map[int32]*field
}

// Initialise the server
func (server *ChannelServer) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work
	server.migrating = make(map[mnet.Client]byte)
	server.players = make(map[mnet.Client]*player)

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

	server.fields = make(map[int32]*field)

	for fieldID, nxMap := range nx.GetMaps() {

		server.fields[fieldID] = &field{
			id:     fieldID,
			data:   nxMap,
			server: server,
		}

		server.fields[fieldID].calculateFieldLimits()
		server.fields[fieldID].createInstance()
	}

	log.Println("Initialised game state")
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
	player := server.players[conn]
	char := player.char
	server.fields[char.mapID].removePlayer(conn, player.instanceID)
	delete(server.players, conn)

	if _, ok := server.migrating[conn]; ok {
		delete(server.migrating, conn)
	} else {
		_, err := server.db.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}

	conn.Cleanup()
}
