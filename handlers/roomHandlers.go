package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/character"
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
			fmt.Println("Create Room type 1")
		case 2:
			fmt.Println("Create Room type 2")
		case 3:
			// create trade
			newTradeRoom := channel.Room{Type: 0x03}

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				newTradeRoom.Participants[0] = char
				displayInfo := []character.Character{char.Character}
				char.SendPacket(packets.RoomShowTradeWindow(0, displayInfo))
			})

			channel.ActiveRooms.Add(newTradeRoom)
		}

	case 0x01:
		fmt.Println("case 1", reader)
	case 0x02:
		// send invite
		charID := reader.ReadInt32()

		channel.Players.OnCharacterFromID(charID, func(recipient *channel.MapleCharacter) {
			channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
				for _, p := range r.Participants {
					if p == nil {
						// we have free space therefore send invite
						channel.Players.OnCharacterFromConn(conn, func(sender *channel.MapleCharacter) {
							recipient.SendPacket(packets.RoomInvite(r.Type, sender.GetName(), r.ID))
						})
					}
				}
			})
		})
	case 0x03:
		//reject
		roomID := reader.ReadInt32()
		rejectCode := reader.ReadByte()

		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			if len(r.Participants) > 0 {
				channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
					r.Participants[0].SendPacket(packets.RoomInviteResult(rejectCode, recipient.GetName()))
				})
			}
		})
	case 0x04:
		//accept
		roomID := reader.ReadInt32()
		// unknown := reader.ReadByte()

		activeRoom := false

		channel.ActiveRooms.OnID(roomID, func(r *channel.Room) {
			activeRoom = true

			displayInfo := []character.Character{}
			spaceAvailable := false
			roomPos := 1

			for i, p := range r.Participants { // change this to walk over the max size
				if p == nil {
					channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
						p = recipient
					})
					spaceAvailable = true
					roomPos = i
					displayInfo = append(displayInfo, p.Character)
				} else {
					displayInfo = append(displayInfo, p.Character)
				}
			}

			if !spaceAvailable {
				channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
					recipient.SendPacket(packets.RoomFull())
				})
			} else {
				channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
					recipient.SendPacket(packets.RoomShowTradeWindow(byte(roomPos), displayInfo))
				})

				for i, p := range r.Participants {
					if p != nil && i != roomPos {

					}
				}
			}
		})

		if !activeRoom {
			channel.Players.OnCharacterFromConn(conn, func(recipient *channel.MapleCharacter) {
				recipient.SendPacket(packets.RoomClosed())
			})
		}

	case 0x0A:
		// close window
		roomID := int32(-1)

		channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
			for i, p := range r.Participants {
				if p.GetConn() == conn {
					r.Participants[i] = nil
					break
				}
			}

			for _, p := range r.Participants {
				if p != nil {
					return
				}
			}

			roomID = r.ID
		})

		channel.ActiveRooms.Remove(roomID)

	default:
		fmt.Println("Unkown case type", operation, reader)
	}
}
