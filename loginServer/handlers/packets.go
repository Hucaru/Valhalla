package handlers

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func RequestWorlds(isAdmin bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.LOGIN_REQUEST_WORLD_SUMMARY)
	if isAdmin {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}

	return p
}

func SendWorlds(worlds []worldData) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(byte(len(worlds)))

	for i := len(worlds) - 1; i > -1; i-- {
		var numberOfValidChannels byte

		for j := range worlds[i].Channels {
			if worlds[i].Channels[j].ID != 0xFF {
				numberOfValidChannels++
			}
		}

		pac := gopacket.NewPacket()
		pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
		pac.WriteByte(worlds[i].ID)     // world id
		pac.WriteString(worlds[i].Name) // World name -
		pac.WriteByte(worlds[i].Ribbon) // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
		pac.WriteString(worlds[i].Event)
		pac.WriteByte(0)                           // ? exp event notification?
		pac.WriteByte(byte(numberOfValidChannels)) // number of channels

		for j, channel := range worlds[i].Channels {
			if channel.ID == 0xFF {
				continue
			}
			pac.WriteString(channel.Name)      // channel name
			pac.WriteInt32(channel.Population) // Population
			pac.WriteByte(worlds[i].ID)        // world id
			pac.WriteByte(channel.ID)          // channel id
			pac.WriteByte(byte(j - 1))         //?
		}

		p.WriteInt16(int16(pac.Size()))
		p.Append(pac)
	}

	return p
}

func RequestWorldInfo(worldID int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.LOGIN_REQUEST_WORLD_STATUS)
	p.WriteInt16(worldID)

	return p
}

func SendWorldInfo(warning byte, population byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_LOGIN_WORLD_META)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}

func RequestMigrationInfo(worldID uint32, channelID byte, charId int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.LOGIN_REQUEST_MIGRATION_INFO)
	p.WriteByte(byte(worldID))
	p.WriteByte(byte(channelID))
	p.WriteInt32(charId)

	return p
}

func SendMigrationInfo(ip []byte, port uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteBytes(ip)
	p.WriteUint16(port)

	return p
}

func sendMigrationToChan(ip []byte, port uint16, charID int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteBytes(ip)
	p.WriteUint16(port)
	p.WriteInt32(charID)

	return p
}
