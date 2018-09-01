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
		roomType := reader.ReadByte()

		switch roomType {
		case 0:
			fmt.Println("Create Room type 0")
		case 1:
			// create memory game
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadByte() == 0x01 {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				channel.CreateOmokGame(char, name, password, boardType)
			})
		case 2:
			// create memory game
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadByte() == 0x01 {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				channel.CreateMemoryGame(char, name, password, boardType)
			})
		case 3:
			// create trade
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				channel.CreateTradeRoom(char)
			})
		case 4:
			// create personal shop
		case 5:
			// create other shop
		default:
			fmt.Println("Unknown room", roomType, reader)
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
					recipient.SendPacket(packets.RoomInvite(r.RoomType, sender.GetName(), r.ID))
				})
			})
		})
	case 0x03:
		//reject
		roomID := reader.ReadInt32()
		rejectCode := reader.ReadByte()

		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				r.Broadcast(packets.RoomInviteResult(rejectCode, recipient.GetName())) // I think we can broadcast this to everyone

				if r.RoomType == 0x03 {
					// Can't remember if a reject caused the window cancel in original
					r.Broadcast(packets.RoomLeave(0, 2))
				}
			})
		})
	case 0x04:
		//accept
		roomID := reader.ReadInt32()
		hasPassword := false
		var password string

		if reader.ReadByte() > 0 {
			hasPassword = true
			password = reader.ReadString(int(reader.ReadInt16()))
		}

		activeRoom := false
		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			activeRoom = true

			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				if hasPassword {
					if password != r.GetPassword() {
						recipient.SendPacket(packets.RoomIncorrectPassword())
						return
					}
				}

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
	case 0x2A:
		// Request tie
	case 0x2C:
		// Request give up
	case 0x32:
		// Ready button pressed
		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.Broadcast(packets.RoomReady())
		})
	case 0x30:
		// Request exit during game
	case 0x33:
		// Unready
		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.Broadcast(packets.RoomUnReady())
		})
	case 0x34:
		// owner expells
		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				r.RemoveParticipant(r.GetParticipantFromSlot(1))
				r.Broadcast(packets.RoomYellowChat(0, char.GetName())) // sending this causes a crash to login screen when re-join
			})
		})
	case 0x35:
		// Game start
		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.InProgress = true

			if p, valid := r.GetBox(); valid {
				channel.Maps.GetMap(r.MapID).SendPacket(p)
			}

			if r.RoomType == 0x01 {
				r.Broadcast(packets.RoomOmokStart(r.P1Turn))
			} else if r.RoomType == 0x02 {

			}

		})
	case 0x37:
		// change turn
		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.Broadcast(packets.RoomOmokSkip(r.P1Turn))
			r.P1Turn = !r.P1Turn
		})
	case 0x38:
		// place piece
		x := reader.ReadInt32()
		y := reader.ReadInt32()
		piece := reader.ReadByte()

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			r.PlacePiece(x, y, piece)
		})

	default:
		fmt.Println("Unkown case type", operation, reader)
	}
}
