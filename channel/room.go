package channel

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type roomer interface {
	id() int32
	setID(int32)
	addPlayer(*Player) bool
	closed() bool
	present(int32) bool
	chatMsg(*Player, string)
	ownerID() int32
}

type room struct {
	roomID        int32
	ownerPlayerID int32
	roomType      byte
	players       []*Player
}

type boxDisplayer interface {
	displayBytes() []byte
}

func (r room) id() int32 {
	return r.roomID
}

func (r *room) setID(id int32) {
	r.roomID = id
}

func (r room) ownerID() int32 {
	return r.ownerPlayerID
}

func (r *room) addPlayer(plr *Player) bool {
	if len(r.players) == 0 {
		r.ownerPlayerID = plr.ID
	} else if len(r.players) == constant.RoomMaxPlayers {
		plr.Send(packetRoomFull())
		return false
	}

	for _, v := range r.players {
		if v == plr {
			return false
		}
	}

	r.players = append(r.players, plr)

	return true
}

func (r *room) removePlayer(plr *Player) bool {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			r.players = append(r.players[:i], r.players[i+1:]...) // preserve order for slot numbers
			return true
		}
	}

	return false
}

func (r room) send(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r room) sendExcept(p mpacket.Packet, plr *Player) {
	for _, v := range r.players {
		if v.Conn == plr.Conn {
			continue
		}
		v.Send(p)
	}
}

func (r *room) sendToOwner(p mpacket.Packet) {
	for _, v := range r.players {
		if v.ID == r.ownerPlayerID {
			v.Send(p)
		}
	}
}

func (r room) closed() bool {
	return len(r.players) == 0
}

func (r room) chatMsg(plr *Player, msg string) {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			r.send(packetRoomChat(plr.Name, msg, byte(i)))
		}
	}
}

func (r room) present(id int32) bool {
	for _, v := range r.players {
		if v.ID == id {
			return true
		}
	}

	return false
}

func packetRoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, players []*Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowWindow)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, v := range players {
		p.WriteByte(byte(i))
		p.Append(v.displayBytes())
		if roomType != constant.MiniRoomTypeTrade && roomType != constant.MiniRoomTypePlayerShop {
			p.WriteInt32(0) // games only - memory card game seed? board settings?
		}
		p.WriteString(v.Name)
	}

	p.WriteByte(constant.RoomPacketEndList)

	if roomType == constant.MiniRoomTypeTrade {
		return p
	}

	if roomType == constant.MiniRoomTypePlayerShop {
		p.WriteString(roomTitle)
		p.WriteByte(constant.RoomShopItemListUnknown)
		p.WriteByte(0)
		return p
	}

	for i, v := range players {
		p.WriteByte(byte(i))
		p.WriteInt32(0) // not sure what this is!?
		p.WriteInt32(v.miniGameWins)
		p.WriteInt32(v.miniGameDraw)
		p.WriteInt32(v.miniGameLoss)
		p.WriteInt32(v.miniGamePoints)
	}

	p.WriteByte(constant.RoomPacketEndList)
	p.WriteString(roomTitle)
	p.WriteByte(boardType)
	p.WriteByte(0)

	return p
}

func packetRoomJoin(roomType, roomSlot byte, plr *Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketJoin)
	p.WriteByte(roomSlot)
	p.Append(plr.displayBytes())
	p.WriteString(plr.Name)

	if roomType == constant.MiniRoomTypeTrade || roomType == constant.MiniRoomTypePlayerShop {
		return p
	}

	p.WriteInt32(1) // not sure what this is!?
	p.WriteInt32(plr.miniGameWins)
	p.WriteInt32(plr.miniGameDraw)
	p.WriteInt32(plr.miniGameLoss)
	p.WriteInt32(plr.miniGamePoints)

	return p
}

func packetRoomLeave(roomSlot byte, leaveCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketLeave)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode) // 2 - trade cancelled, 6 - trade success

	return p
}

func packetRoomYellowChat(msgType byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomChat)
	p.WriteByte(constant.RoomChatTypeNotice)
	// expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4,
	// called to leave/expeled: 5, cancelled leave: 6, entered: 7, can't start lack of mesos:8
	// has matched cards: 9
	p.WriteByte(msgType)
	p.WriteString(name)

	return p
}

func packetRoomReady() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomReadyButtonPressed)

	return p
}

func packetRoomChat(sender, message string, roomSlot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomChat)
	p.WriteByte(constant.RoomChatTypeChat)
	p.WriteByte(roomSlot)
	p.WriteString(sender + " : " + message)

	return p
}

func packetRoomUnready() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomUnready)

	return p
}

func packetRoomInvite(roomType byte, name string, roomID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketInvite)
	p.WriteByte(roomType)
	p.WriteString(name)
	p.WriteInt32(roomID)

	return p
}

func packetRoomInviteResult(resultCode byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketInviteResult)
	p.WriteByte(resultCode)
	p.WriteString(name)

	return p
}

func packetRoomShowAccept() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowAccept)

	return p
}

func packetRoomEnterErrorMsg(errorCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowWindow)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
}

func packetRoomClosed() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterClosed)
}

func packetRoomFull() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterFull)
}

func packetRoomBusy() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterBusy)
}

func packetRoomNotAllowedWhenDead() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNotAllowedDead)
}

func packetRoomNotAllowedDuringEvent() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNotAllowedEvent)
}

func packetRoomThisCharacterNotAllowed() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterThisCharNotAllow)
}

func packetRoomNoTradeAtm() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNoTradeATM)
}

func packetRoomMiniRoomNotHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x08)
}

func packetRoomcannotCreateMiniroomHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterCannotCreateHere)
}

func packetRoomCannotStartGameHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterCannotStartHere)
}

func packetRoomPersonalStoreFMOnly() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterStoreFMOnly)
}
func packetRoomGarbageMsgAboutFloorInFm() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterGarbageFloorFM)
}

func packetRoomMayNotEnterStore() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterMayNotEnterStore)
}

func packetRoomCannotEnterTournament() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterUnableEnterTournament)
}

func packetRoomGarbageTradeMsg() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterGarbageTradeMsg)
}

func packetRoomNotEnoughMesos() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterNotEnoughMesos)
}
