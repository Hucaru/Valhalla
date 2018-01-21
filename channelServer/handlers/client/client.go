package client

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"math"
	"strings"

	"github.com/Hucaru/Valhalla/channelServer/handlers/login"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func HandlePacket(conn *Connection, reader gopacket.Reader) {
	opcode := reader.ReadByte()

	switch opcode {
	case constants.RECV_CHANNEL_PLAYER_LOAD:
		handlePlayerLoad(reader, conn)
	case constants.RECV_CHANNEL_MOVEMENT:
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		handlePlayerSendAllChat(reader, conn)
	case constants.RECV_CHANNEL_ADD_BUDDY:
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}

func handlePlayerSendAllChat(reader gopacket.Reader, conn *Connection) {
	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.isAdmin {
		command := strings.SplitN(msg[ind+1:], " ", -1)
		switch command[0] {
		case "packet":
			packet := string(command[1])
			data, err := hex.DecodeString(packet)

			if err != nil {
				log.Println("Eror in decoding string for gm command packet:", packet)
				break
			}
			log.Println("Sent packet:", hex.EncodeToString(data))
			conn.Write(data)
		default:
			log.Println("Unkown GM command", command)
		}

	}
}

func handlePlayerLoad(reader gopacket.Reader, conn *Connection) {
	charID := reader.ReadUint32() // validate this and net address from the migration packet

	if !login.ValidateMigration(charID) {
		log.Println("Invalid migration char id:", charID)
		conn.Close()
	}

	char := character.GetCharacter(charID)

	conn.SetAdmin(true)

	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	pac.WriteInt32(0) // Channel ID
	pac.WriteByte(1)  // 0 portals
	pac.WriteByte(1)  // Is connecting

	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Println(err.Error())
		return
	}
	pac.WriteBytes(randomBytes)
	pac.WriteBytes(randomBytes)
	pac.WriteBytes(randomBytes)
	pac.WriteBytes(randomBytes)
	pac.WriteBytes([]byte{0xFF, 0xFF})   // ??
	pac.WriteUint32(charID)              // charid
	pac.WritePaddedString(char.Name, 13) // name
	pac.WriteByte(char.Gender)           // Gender
	pac.WriteByte(char.Skin)             // Skin
	pac.WriteUint32(char.Face)           // Face
	pac.WriteUint32(char.Hair)           // Hair

	pac.WriteInt64(0) // Pet Cash ID

	pac.WriteByte(char.Level)   // Level
	pac.WriteUint16(char.Job)   // Jobid
	pac.WriteUint16(char.Str)   //charc.str
	pac.WriteUint16(char.Dex)   //charc.dex
	pac.WriteUint16(char.Int)   //charc.intt
	pac.WriteUint16(char.Luk)   //charc.luk
	pac.WriteUint16(char.HP)    //charc.hp);
	pac.WriteUint16(char.MaxHP) //charc.mhp //Needs to be set to Original MAX HP before using hyperbody.
	pac.WriteUint16(char.MP)    //charc.mp
	pac.WriteUint16(char.MaxMP) //charc.mmp
	pac.WriteUint16(char.AP)    //charc.ap
	pac.WriteUint16(char.SP)    //charc.sp
	pac.WriteUint32(char.EXP)   //charc.exp
	pac.WriteUint16(char.Fame)  //charc.fame

	pac.WriteUint32(char.CurrentMap)  //definitly map ID
	pac.WriteByte(char.CurrentMapPos) // map pos

	pac.WriteByte(5) // budy list size?
	pac.WriteUint32(char.Mesos)

	pac.WriteByte(255) // Equip inv size
	pac.WriteByte(255) // User inv size
	pac.WriteByte(255) // Setup inv size
	pac.WriteByte(255) // Etc inv size
	pac.WriteByte(255) // Cash inv size

	char.Items = character.GetCharacterItems(charID)

	// Equips -50 -> -1 normal equips
	// Cash items / equip covers -150 to -101 maybe?

	for _, v := range char.Items {
		if v.SlotID < 0 {
			// Equips
			pac.WriteByte(byte(math.Abs(float64(v.SlotID))))
			pac.WriteByte(byte(v.ItemID / 1000000))
			pac.WriteUint32(v.ItemID)
			pac.WriteByte(0) // not a cash item, switch to 1 if it is
			pac.WriteUint64(v.ExpireTime)
			pac.WriteByte(v.UpgradeSlots)
			pac.WriteByte(v.Level)
			pac.WriteUint16(v.Str)
			pac.WriteUint16(v.Dex)
			pac.WriteUint16(v.Intt)
			pac.WriteUint16(v.Luk)
			pac.WriteUint16(v.HP)
			pac.WriteUint16(v.MP)
			pac.WriteUint16(v.Watk)
			pac.WriteUint16(v.Matk)
			pac.WriteUint16(v.Wdef)
			pac.WriteUint16(v.Mdef)
			pac.WriteUint16(v.Accuracy)
			pac.WriteUint16(v.Avoid)
			pac.WriteUint16(v.Hands)
			pac.WriteUint16(v.Speed)
			pac.WriteUint16(v.Jump)
			pac.WriteInt32(0)
		} else {
			// Inventory items
		}
	}

	// Cash items / equip covers -150 to -101

	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)

	conn.Write(pac)
}
