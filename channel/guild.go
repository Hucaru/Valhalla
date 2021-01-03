package channel

import (
	"strconv"

	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
)

// TODO: login server needs to send a deleted character event so that they can leave the guild for playing players

type guild struct {
	players []*player
	internal.Guild

	// id       int32
	// capacity int32
	// notice   string
	// name     string

	// master   string
	// jrMaster string
	// member1  string
	// member2  string
	// member3  string

	// logoBg, logoBgColour, logo int16
	// logoColour                 byte

	// names  []string
	// jobs   []int32
	// levels []int32
	// online []bool
	// ranks  []int32
}

func packetGuildInfo(id int32, name string, memberCount byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if len(name) == 0 {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(1)   // guild id (value cannot be zero)
	p.WriteString(name)

	// 5 ranks each have a title
	p.WriteString("rank1")
	p.WriteString("rank2")
	p.WriteString("rank3")
	p.WriteString("rank4")
	p.WriteString("rank5")

	capacity := 250                  // maximum
	p.WriteByte(byte(capacity) - 10) // member count

	// iterate over all members and output ids
	for i := 0; i < capacity-10; i++ {
		p.WriteInt32(int32(i + 1))
	}

	// iterate over all members and input their info
	for i := 0; i < capacity-10; i++ {
		p.WritePaddedString("Player "+strconv.Itoa(i), 13) // name
		p.WriteInt32(510)                                  // job
		p.WriteInt32(255)                                  // level

		if i > 4 {
			p.WriteInt32(5) // rank starts at 1
		} else {
			p.WriteInt32(int32(i + 1)) // rank starts at 1
		}

		if i%2 == 0 {
			p.WriteInt32(1) // online or not
		} else {
			p.WriteInt32(0)
		}

		p.WriteInt32(int32(i)) // ?
	}

	p.WriteInt32(int32(capacity)) // capacity
	p.WriteInt16(1030)            // logo background
	p.WriteByte(3)                // logo bg colour
	p.WriteInt16(4017)            // logo
	p.WriteByte(2)                // logo colour
	p.WriteString("notice")       // notice
	p.WriteInt32(9999)            // ?

	return p
}

func packetGuildPlayerOnlineNotice(guildID, playerIndex int32, online bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x3d)
	p.WriteInt32(guildID)
	p.WriteInt32(playerIndex)
	p.WriteBool(online)

	return p
}

func packetGuildInviteNotAccepting(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x35)
}

func packetGuildInviteHasAnother(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x36)
}

func packetGuildInviteRejected(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x37)
}

func packetGuildInviteResult(name string, code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(code)
	p.WriteString(name)

	return p
}

/*
0x32 - guildDisbanded npc dialogue box (updates ui)
i32 - guildID

0x3a - guild capacity npc dialogue box (ui not updated)
i32 - guildID
i8 - capacity

0x3c -
i32 - guildID
i32
i32
i32

0x34 - npc dialogue box saying problem has occured during disbandon

0x3b - npc dialogue box saying problem has occured during capacity increase

0x38 - admin cannot make guild message

0x49 -
i32
i32 - amount
for amount:
	name
	i32

0x4a - less than 5 members remaning, guild quest will end in 5 seconds

0x4b - user that registered has disconnected, quest will end in 5 seconds

0x4c - guild quest status and position in queue
i8 - channelID
i32 - position in queue

0x48 -
i32 - guildID
i32 - ?

0x3c -
i32 - guildID
i32
i8

0x3b -
i32 - guildID
name

between 0x3f - 0x47 -
i32 - guildID
i16
i8
i16
i8

0x3e - update rank titles (dialogue box comes up saying it has been saved) ui is updated
i32 - guildID
name  - master
name
name
name
name - member

0x30 - you are not in the guild

0x29 - the guild you are trying to join has reached maximum capacity

*/
