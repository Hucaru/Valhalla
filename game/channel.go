package game

import (
	"database/sql"
	"log"
	"math/rand"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Channel server state
type Channel struct {
	id        byte
	db        *sql.DB
	dispatch  chan func()
	world     mnet.Server
	ip        []byte
	port      int16
	maxPop    int16
	migrating map[mnet.Client]byte
	sessions  map[mnet.Client]*entity.Character
	channels  [20]entity.Channel
}

// Initialise the server
func (server *Channel) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work
	server.migrating = make(map[mnet.Client]byte)
	server.sessions = make(map[mnet.Client]*entity.Character)

	var err error
	server.db, err = sql.Open("mysql", dbuser+":"+dbpassword+"@tcp("+dbaddress+":"+dbport+")/"+dbdatabase)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.db.Ping()

	if err != nil {
		log.Fatal(err.Error()) // change to attempt to re-connect
	}

	log.Println("Connected to database")

}

// RegisterWithWorld server
func (server *Channel) RegisterWithWorld(conn mnet.Server, ip []byte, port int16, maxPop int16) {
	server.world = conn
	server.ip = ip
	server.port = port
	server.maxPop = maxPop

	server.registerWithWorld()
}

func (server *Channel) registerWithWorld() {
	p := mpacket.CreateInternal(opcode.ChannelNew)
	p.WriteBytes(server.ip)
	p.WriteInt16(server.port)
	p.WriteInt16(server.maxPop)
	server.world.Send(p)
}

// HandleServerPacket from world
func (server *Channel) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
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

func (server *Channel) handleNewChannelBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by world server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithWorld()
}

func (server *Channel) handleNewChannelOK(conn mnet.Server, reader mpacket.Reader) {
	server.id = reader.ReadByte()
	log.Println("Registered as channel", server.id)
}

func (server *Channel) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

// ClientConnected to server
func (server *Channel) ClientConnected(conn net.Conn, clientEvent chan *mnet.Event, packetQueueSize int) {
	keySend := [4]byte{}
	rand.Read(keySend[:])
	keyRecv := [4]byte{}
	rand.Read(keyRecv[:])

	client := mnet.NewClient(conn, clientEvent, packetQueueSize, keySend, keyRecv)

	go client.Reader()
	go client.Writer()

	conn.Write(entity.PacketClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
}

// ClientDisconnected from server
func (server *Channel) ClientDisconnected(conn mnet.Client) {
	if _, ok := server.migrating[conn]; ok {
		// conn.GetWorldID()
		// conn.GetChannelID()
		// set migrating channel and world in db
		delete(server.migrating, conn)
	} else if conn.GetLogedIn() {
		_, err := server.db.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}
	conn.Cleanup()
}

// HandleClientPacket
func (server *Channel) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		server.playerConnect(conn, reader)
	case opcode.RecvCHannelChangeChannel:
		server.playerChangeChannel(conn, reader)
	case opcode.RecvChannelUserPortal:
		// server.playerUsePortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
	case opcode.RecvChannelPlayerMovement:
		// server.playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		// server.playerStand(conn, reader)
	case opcode.RecvChannelPlayerUserChair:
		// server.playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		// server.playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		// server.playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		// server.playerMagicSkill(conn, reader)
	case opcode.RecvChannelDmgRecv:
		// server.playerTakeDamage(conn, reader)
	case opcode.RecvChannelPlayerSendAllChat:
		// server.chatSendAll(conn, reader)
	case opcode.RecvChannelSlashCommands:
		// server.chatSlashCommand(conn, reader)
	case opcode.RecvChannelCharacterUIWindow:
		// server.handleUIWindow(conn, reader)
	case opcode.RecvChannelEmote:
		// server.playerEmote(conn, reader)
	case opcode.RecvChannelNpcDialogue:
		// server.npcChatStart(conn, reader)
	case opcode.RecvChannelNpcDialogueContinue:
		// server.npcChatContinue(conn, reader)
	case opcode.RecvChannelNpcShop:
	case opcode.RecvChannelInvMoveItem:
		// server.playerMoveInventoryItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		// server.playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		// server.playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		// server.playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		// server.playerSpecialSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		// server.playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
	case opcode.RecvChannelGuildManagement:
	case opcode.RecvChannelGuildReject:
	case opcode.RecvChannelAddBuddy:
	case opcode.RecvChannelMobControl:
		// server.mobControl(conn, reader)
	case opcode.RecvChannelNpcMovement:
		// server.npcMovement(conn, reader)
	default:
		log.Println("UNKNOWN CLIENT PACKET:", reader)
	}
}
