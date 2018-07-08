package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/movement"
	"github.com/Hucaru/Valhalla/npcdialogue"
	"github.com/Hucaru/Valhalla/packets"
)

func handlePlayerConnect(conn interop.ClientConn, reader maplepacket.Reader) {
	charID := reader.ReadInt32()

	char := character.GetCharacter(charID)
	char.SetItems(character.GetCharacterItems(char.GetCharID()))
	char.SetSkills(character.GetCharacterSkills(char.GetCharID()))

	var isAdmin bool
	err := connection.Db.QueryRow("SELECT isAdmin from users where userID=?", char.GetUserID()).Scan(&isAdmin)

	if err != nil {
		panic(err)
	}

	var channelID int32 = 0 // Get from world server or docker-compose

	conn.SetAdmin(isAdmin)
	conn.SetIsLoggedIn(true) // review if this is needed
	conn.SetChanID(channelID)

	channel.Players.AddPlayer(conn, &char)

	conn.AddCloseCallback(func() {
		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			err := character.SaveCharacter(&char.Character)

			if err != nil {
				log.Println("Unable to save character data")
			}
			channel.Maps.GetMap(char.GetCurrentMap()).RemovePlayer(conn)
		})

		npcdialogue.RemoveSession(conn)

		channel.Players.RemovePlayer(conn)
	})

	conn.Write(packets.PlayerEnterGame(char, channelID))

	portal := channel.Maps.GetMap(char.GetCurrentMap()).GetPortals()[char.GetCurrentMapPos()]

	char.SetX(portal.GetX())
	char.SetY(portal.GetY())

	channel.Maps.GetMap(char.GetCurrentMap()).AddPlayer(conn)

	conn.Write(packets.MessageScrollingHeader(channel.GetHeader()))

	// Send party info

	// Send guild info

}

func handleTakeDamage(conn interop.ClientConn, reader maplepacket.Reader) {
	dmgType := reader.ReadByte()
	ammount := reader.ReadInt32()

	var mobID int32
	var reduction byte
	var stance byte
	var hit byte

	switch dmgType {
	case 0xFE: // map or fall damage
	default:
		mobID = reader.ReadInt32()
		reader.ReadInt32() // some form of map object id?
		hit = reader.ReadByte()
		reduction = reader.ReadByte()
		stance = reader.ReadByte()
	}

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		//char.TakeDamage(ammount)

		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.PlayerReceivedDmg(char.GetCharID(),
			ammount, dmgType, mobID, hit, reduction, stance),
			conn)
	})
}

func handleRequestAvatarInfoWindow(conn interop.ClientConn, reader maplepacket.Reader) {
	charID := reader.ReadInt32()

	channel.Players.OnCharacterFromID(charID, func(char *channel.MapleCharacter) {
		conn.Write(packets.PlayerAvatarSummaryWindow(charID, char.Character, "Admins"))
	})
}

func handlePassiveRegen(conn interop.ClientConn, reader maplepacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		if char.GetHP() == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
			return
		}

		if hp > 0 {
			char.SetHP(hp)
		} else if mp > 0 {
			char.SetMP(mp)
		}
	})

	// If in party return id and new hp, then update hp bar for party members
}

func handleChangeStat(conn interop.ClientConn, reader maplepacket.Reader) {
	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		if char.GetAP() == 0 {
			return
		}

		stat := reader.ReadInt32()

		switch stat {
		case constants.STR_ID:
			char.SetStr(char.GetStr() + 1)
		case constants.DEX_ID:
			char.SetDex(char.GetDex() + 1)
		case constants.INT_ID:
			char.SetInt(char.GetInt() + 1)
		case constants.LUK_ID:
			char.SetLuk(char.GetLuk() + 1)
		case constants.MAX_HP_ID:
			char.SetMaxHP(char.GetMaxHP() + 1)
		case constants.MAX_MP_ID:
			char.SetMaxMP(char.GetMaxMP() + 1)
		default:
			log.Println("Unknown stat ID:", stat)
		}
	})
}

func handleUpdateSkillRecord(conn interop.ClientConn, reader maplepacket.Reader) {
	skillID := reader.ReadInt32()
	newLevel := int32(0)

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {

		newSP := char.GetSP() - 1
		char.SetSP(newSP)

		skills := char.GetSkills()

		if _, exists := skills[skillID]; exists {
			newLevel = skills[skillID] + 1
		} else {
			newLevel = 1
		}

		char.UpdateSkill(skillID, newLevel)
	})
}

func handlePlayerMovement(conn interop.ClientConn, reader maplepacket.Reader) {
	reader.ReadBytes(5) // used in movement validation
	nFrags := reader.ReadByte()

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		movement.ParseFragments(nFrags, char, reader)
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.PlayerMove(char.GetCharID(), reader.GetBuffer()[2:]), conn)
	})
}

func handlePlayerEmoticon(conn interop.ClientConn, reader maplepacket.Reader) {
	emoticon := reader.ReadInt32()
	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.PlayerEmoticon(char.GetCharID(), emoticon), conn)
	})
}
