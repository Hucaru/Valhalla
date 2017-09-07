package client

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func HandlePacket(conn *Connection, reader gopacket.Reader) {
	opcode := reader.ReadByte()

	switch opcode {
	case constants.RECV_CHANNEL_CLIENT_MIGRATION:
		handleServerJoin(reader, conn)
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

func handleServerJoin(reader gopacket.Reader, conn *Connection) {
	log.Println(reader.GetBuffer())

	pac := gopacket.NewPacket()
	pac.WriteByte(0x0B)
	pac.WriteByte(1)
	pac.WriteBytes([]byte{192, 168, 1, 117})
	pac.WriteInt16(8684)
	conn.Write(pac)
}

func handlePlayerLoad(reader gopacket.Reader, conn *Connection) {

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
	pac.WriteInt32(8)                    // charid
	pac.WriteBytes([]byte("[GM]Hucaru")) // name
	pac.WriteBytes([]byte{0, 0, 0})      //name padding
	pac.WriteByte(0)                     // Gender
	pac.WriteByte(0)                     // Skin
	pac.WriteInt32(20000)                // Face
	pac.WriteInt32(30020)                // Hair

	pac.WriteInt64(0) // Pet Cash ID

	pac.WriteByte(1)    // Level
	pac.WriteInt16(0)   // Jobid
	pac.WriteInt16(7)   //charc.str
	pac.WriteInt16(5)   //charc.dex
	pac.WriteInt16(6)   //charc.intt
	pac.WriteInt16(7)   //charc.luk
	pac.WriteInt16(100) //charc.hp);
	pac.WriteInt16(100) //charc.mhp //Needs to be set to Original MAX HP before using hyperbody.
	pac.WriteInt16(50)  //charc.mp
	pac.WriteInt16(50)  //charc.mmp
	pac.WriteInt16(0)   //charc.ap
	pac.WriteInt16(0)   //charc.sp
	pac.WriteInt32(0)   //charc.exp
	pac.WriteInt16(0)   //charc.fame

	pac.WriteInt32(100000000) //definitly map ID
	pac.WriteByte(0)          // map pos

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
