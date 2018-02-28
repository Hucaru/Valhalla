package server

import "github.com/Hucaru/gopacket"

type mapleNpc struct {
	npcID     uint32
	spawnID   uint32
	spawnX    int16
	spawnY    int16
	spawnFh   int16
	x         int16
	y         int16
	faceRight bool
}

// Returns the spawn packet for the selected NPC
func (this *mapleNpc) Spawn() gopacket.Packet {
	return this.Show()
}

// Returns the show packet for the selected NPC
func (this *mapleNpc) Show() gopacket.Packet {
	p := gopacket.NewPacket()

	return p
}
