package server

import (
	"github.com/Hucaru/Valhalla/channelServer/client"
	"github.com/Hucaru/gopacket"
)

type mapleMob struct {
	monsterID uint32
	spawnID   uint32
	spawnX    int16
	spawnY    int16
	spawnFh   int16
	x         int16
	y         int16
	faceRight bool

	controller *client.Conn

	boss     bool
	accuracy uint16
	exp      uint32
	level    byte
	maxHp    uint16
	hp       uint16
	maxMp    uint16
	mp       uint16
}

// Returns the spawn packet for the selected mob
func (this *mapleMob) Spawn() gopacket.Packet {
	p := gopacket.NewPacket()

	return p
}

// Returns the show packet for the selected mob
func (this *mapleMob) Show() gopacket.Packet {
	p := gopacket.NewPacket()

	return p
}
