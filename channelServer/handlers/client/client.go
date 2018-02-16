package client

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/channelServer/handlers/client/packets"
	"github.com/Hucaru/Valhalla/channelServer/handlers/client/skill"
	"github.com/Hucaru/Valhalla/channelServer/handlers/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
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

	if !validateNewConnection(charID) {
		conn.Close()
	}

	_, channelID := world.GetAssignedIDs()

	char := character.GetCharacter(charID)

	conn.SetAdmin(true)

	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	pac.WriteUint32(uint32(channelID)) // Channel ID
	pac.WriteByte(1)                   // 0 portals
	pac.WriteByte(1)                   // Is connecting

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
	pac.WriteBytes([]byte{0xFF, 0xFF})   // seperators? For what?
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
	pac.WriteUint16(char.Intt)  //charc.intt
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

	pac.WriteByte(20) // budy list size?
	pac.WriteUint32(char.Mesos)

	pac.WriteByte(char.EquipSlotSize) // Equip inv size
	pac.WriteByte(char.UsetSlotSize)  // User inv size
	pac.WriteByte(char.SetupSlotSize) // Setup inv size
	pac.WriteByte(char.EtcSlotSize)   // Etc inv size
	pac.WriteByte(char.CashSlotSize)  // Cash inv size

	char.Equips = character.GetCharacterItems(charID)

	// Equips -50 -> -1 normal equips
	for _, v := range char.Equips {
		if v.SlotID < 0 && v.SlotID > -20 {
			pac.WriteBytes(packets.AddEquip(v))
		}
	}

	pac.WriteByte(0)

	// Cash item equip covers -150 to -101 maybe?
	for _, v := range char.Equips {
		if v.SlotID < -100 {
			pac.WriteBytes(packets.AddEquip(v))
		}
	}

	pac.WriteByte(0)

	for _, v := range char.Equips {
		if v.SlotID > -1 {
			pac.WriteBytes(packets.AddEquip(v)) // there is a caveat for adding a cash item
		}
	}

	pac.WriteByte(0)

	// use
	pac.WriteByte(1)         // slot id
	pac.WriteByte(2)         // type of item e.g. equip, has amount, cash
	pac.WriteUint32(2070006) //  itemID
	pac.WriteByte(0)
	pac.WriteUint64(0)   // expiration
	pac.WriteUint16(200) // amount
	pac.WriteUint16(0)   // string with name of creator
	pac.WriteUint16(0)   // is it sealed

	// use
	pac.WriteByte(2)         // slot id
	pac.WriteByte(2)         // type of item
	pac.WriteUint32(2000003) //  itemID
	pac.WriteByte(0)
	pac.WriteUint64(0)
	pac.WriteUint16(200) // amount
	pac.WriteUint16(0)
	pac.WriteUint16(0) // is it sealed

	pac.WriteByte(0) // Inventory tab move forward swap

	pac.WriteByte(1)         // slot id
	pac.WriteByte(2)         // type of item
	pac.WriteUint32(3010000) //  itemID
	pac.WriteByte(0)
	pac.WriteUint64(0)
	pac.WriteUint16(1) // amount
	pac.WriteUint16(0)
	pac.WriteUint16(0) // is it sealed

	pac.WriteByte(0) // Inventory tab move forward swap

	// etc
	pac.WriteByte(1)         // slot id
	pac.WriteByte(2)         // type of item
	pac.WriteUint32(4000000) //  itemID
	pac.WriteByte(0)
	pac.WriteUint64(0)
	pac.WriteUint16(200) // amount
	pac.WriteUint16(0)
	pac.WriteUint16(0) // is it sealed

	pac.WriteByte(0) // Inventory tab move forward swap

	// cash pet item :( not working atm
	pac.WriteByte(1)         // slot id
	pac.WriteByte(2)         // Type of item (1 means it is an equip, 2 means inv?, 3 means ?)
	pac.WriteUint32(5000004) //  itemID
	pac.WriteByte(0)
	// pac.WriteUint32(5000004)
	pac.WriteUint64(0)
	pac.WriteUint16(1) // amount
	pac.WriteUint16(0)
	pac.WriteUint16(0) // is it sealed

	pac.WriteByte(0)

	// Skills
	skills := skill.GetCharacterSkills(charID)
	pac.WriteUint16(uint16(len(skills))) // number of skills

	for _, v := range skills {
		pac.WriteUint32(v.SkillID)
		pac.WriteUint32(uint32(v.Level))
	}

	// Quests
	pac.WriteUint16(0) // # of quests

	// Minigame
	pac.WriteUint16(0)
	pac.WriteUint32(0)
	pac.WriteUint32(0)
	pac.WriteUint32(0)
	pac.WriteUint32(0)
	pac.WriteUint32(0)

	pac.WriteUint64(0)
	pac.WriteUint64(0)
	pac.WriteUint64(0)
	pac.WriteUint64(0)
	pac.WriteUint64(0)

	pac.WriteInt64(time.Now().Unix())

	conn.Write(pac)

	// npc spawn
	life := nx.Maps[char.CurrentMap].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(packets.SpawnNPC(uint32(i), v))
		}
	}
}

func validateNewConnection(charID uint32) bool {
	var migratingWorldID, migratingChannelID int8
	err := connection.Db.QueryRow("SELECT isMigratingWorld,isMigratingChannel FROM characters where id=?", charID).Scan(&migratingWorldID, &migratingChannelID)

	if err != nil {
		panic(err.Error())
	}

	if migratingWorldID < 0 || migratingChannelID < 0 {

		return false
	}

	msg := make(chan gopacket.Packet)
	world.InterServer <- connection.NewMessage([]byte{constants.CHANNEL_GET_INTERNAL_IDS}, msg)
	result := <-msg
	r := gopacket.NewReader(&result)

	if r.ReadByte() != byte(migratingWorldID) && r.ReadByte() != byte(migratingChannelID) {
		log.Println("Received invalid migration info for character", charID, "remote hacking")
		records, err := connection.Db.Query("UPDATE characters set migratingWorldID=?, migratingChannelID=? WHERE id=?", -1, -1, charID)

		defer records.Close()

		if err != nil {
			panic(err.Error())
		}

		return false
	}

	return true
}
