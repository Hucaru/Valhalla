package login

import (
	"strconv"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

var worldNames = [...]string{"Scania", "Bera", "Broa", "Windia", "Khaini", "Bellocan", "Mardia", "Kradia", "Yellonde", "Demethos", "Galicia", "El Nido", "Zenith", "Arcania", "Chaos", "Nova", "Renegates"}

func worldListing(worldIndex byte) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(worldIndex)               // world id
	pac.WriteString(worldNames[worldIndex]) // World name -
	pac.WriteByte(3)                        // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
	pac.WriteString("test")
	pac.WriteByte(0)  // ? exp event notification?
	pac.WriteByte(20) // number of channels

	maxPopulation := 150
	population := 50

	for j := 1; j < 21; j++ {
		pac.WriteString(worldNames[worldIndex] + "-" + strconv.Itoa(j))                // channel name
		pac.WriteInt32(int32(1200.0 * (float64(population) / float64(maxPopulation)))) // Population
		pac.WriteByte(worldIndex)                                                      // world id
		pac.WriteByte(byte(j))                                                         // channel id
		pac.WriteByte(byte(j - 1))                                                     //?
	}

	return pac
}

func endWorldList() maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(0xFF)

	return pac
}

func worldInfo(warning byte, population byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_LOGIN_WORLD_META)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}
