package game

import (
	"database/sql"
	"log"
	"math/rand"
	"net"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Channel server state
type Channel struct {
	// maps
	// players
	migrating map[mnet.Client]bool
	db        *sql.DB
	dispatch  chan func()
}

// Initialise the server
func (server *Channel) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work

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
	if isMigrating, ok := server.migrating[conn]; ok && isMigrating {
		// set migrating channel and world in db
		// conn.GetWorldID()
		// conn.GetChannelID()
	} else if conn.GetLogedIn() {
		_, err := server.db.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}
	conn.Cleanup()
	conn.Cleanup()
}

// HandleClientPacket
func (server *Channel) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch mpacket.Opcode(reader.ReadByte()) {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		server.playerConnect(conn, reader)
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
		log.Println("Unkown packet:", reader)
	}
}

func (server *Channel) playerConnect(conn mnet.Client, reader mpacket.Reader) {

}

// HandleServerPacket
func (server *Channel) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {

}
