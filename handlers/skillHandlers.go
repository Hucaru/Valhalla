package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleStandardSkill(conn interop.ClientConn, reader maplepacket.Reader) {
	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32()

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt16() // ?

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.SkillMelee(char.GetCharID(), skillID, targets, hits, display, animation, damages),
			conn)
	})
}

func handleRangedSkill(conn interop.ClientConn, reader maplepacket.Reader) {
	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32() // ?
	invPos := reader.ReadByte()
	reader.ReadBytes(4)
	fmt.Println("Ranged weapon inventory location:", invPos)

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadBytes(14)

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	// hard coded ilbi for now
	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.SkillRanged(char.GetCharID(), skillID, 2070006, targets, hits, display, animation, damages),
			conn)
	})
}

func handleMagicSkill(conn interop.ClientConn, reader maplepacket.Reader) {
	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32() //?

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt16() // ?

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.SkillMagic(char.GetCharID(), skillID, targets, hits, display, animation, damages),
			conn)
	})
}

func handleSpecialSkill(conn interop.ClientConn, reader maplepacket.Reader) {
	skillID := reader.ReadUint32()
	level := reader.ReadByte()

	// add all the various skills that fall under this category
	switch skillID {

	// GM SKILLS
	case 5001000: // gm haste normal
	case 5001001: // gm super dragon roar
	case 5001002: // gm teleport
	case 5101000: // gm heal + dispel
	case 5101001: // gm super haste
	case 5101002: // gm holy symbol
	case 5101003: // gm bless
	case 5101004: // gm hide
		conn.Write(packets.SkillGmHide(true))
	case 5101005: // gm resurect
	default:
		fmt.Println("Unkown skill id:", skillID)
	}

	conn.Write(packets.PlayerStatNoChange()) // Needs a continue packet of some kind?

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.SkillAnimation(char.GetCharID(), skillID, level),
			conn)
	})
}
