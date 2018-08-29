package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleUIWindow(conn *connection.Channel, reader maplepacket.Reader) {
	operation := reader.ReadByte() // Trade operation

	switch operation {
	case 0x00:
		// check not in a room already
		alreadyInRoom := false

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			alreadyInRoom = true
		})

		if alreadyInRoom {
			return
		}

		// create room
		roomType := reader.ReadInt16()

		switch roomType {
		case 0:
			fmt.Println("Create Room type 0")
		case 1:
			fmt.Println("Create Omok Game Room")
		case 2:
			fmt.Println("Create Memory Game Room")
		case 3:
			// create trade
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				channel.CreateTradeRoom(char)
			})
		}

	case 0x01:
		fmt.Println("case 1", reader)
	case 0x02:
		// send invite
		charID := reader.ReadInt32()

		channel.Players.OnCharacterFromID(charID, func(recipient *channel.MapleCharacter) {
			channel.Players.OnCharacterFromConn(conn, func(sender *channel.MapleCharacter) {
				if sender.GetCurrentMap() != recipient.GetCurrentMap() {
					return // hacker
				}

				channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
					recipient.SendPacket(packets.RoomInvite(r.Type, sender.GetName(), r.ID))
				})
			})
		})
	case 0x03:
		//reject - check if window closes for trade
		roomID := reader.ReadInt32()
		rejectCode := reader.ReadByte()

		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				r.Broadcast(packets.RoomInviteResult(rejectCode, recipient.GetName())) // I think we can broadcast this to everyone
			})
		})
	case 0x04:
		//accept
		roomID := reader.ReadInt32()
		// unknown := reader.ReadByte()

		activeRoom := false

		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			activeRoom = true

			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				r.AddParticipant(recipient)
			})
		})

		if !activeRoom {
			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				recipient.SendPacket(packets.RoomClosed())
			})
		}

	case 0x06:
		// chat
		message := reader.ReadString(int(reader.ReadInt16()))
		name := ""
		// roomSlot := byte(0x0)

		channel.Players.OnCharacterFromConn(conn, func(sender *channel.MapleCharacter) {
			name = sender.GetName()
		})

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.SendMessage(name, message)
		})

	case 0x0A:
		// close window
		roomID := int32(-1)
		removeRoom := false

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				removeRoom, roomID = r.RemoveParticipant(char)
			})
		})

		if removeRoom {
			channel.ActiveRooms.Remove(roomID)
		}
	case 0x0D:
		// insert item
		// invTab := reader.ReadByte()
		// itemSlot := reader.ReadInt16()
		// quantity := reader.ReadInt16()
		// tradeWindowSlot := reader.ReadByte()

	case 0x0E:
		// mesos
		// amount := reader.ReadInt32()
	case 0x0F:
		// accept trade button pressed
		removeRoom := false
		roomID := int32(-1)

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				removeRoom, roomID = r.Accept(char)
			})
		})

		if removeRoom {
			channel.ActiveRooms.Remove(roomID)
		}
	default:
		fmt.Println("Unkown case type", operation, reader)
	}
}
