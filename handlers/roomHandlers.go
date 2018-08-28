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
		roomType := reader.ReadInt16()

		switch roomType {
		case 0:
			fmt.Println("Create Room type 0")
		case 1:
			fmt.Println("Create Room type 1")
		case 2:
			fmt.Println("Create Room type 2")
		case 3:
			// check not in a room already
			alreadyInRoom := false

			channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
				alreadyInRoom = true
			})

			if alreadyInRoom {
				return
			}

			// create trade
			newTradeRoom := channel.Room{ID: channel.ActiveRooms.GetNextRoomID()}

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				newTradeRoom.Participants[0] = char
				displayInfo := []character.Character{char.Character}
				char.SendPacket(packets.RoomShowTradeWindow(0, displayInfo))

				channel.ActiveRooms.Add(newTradeRoom)
			})
		}

	case 0x01:
	case 0x02:
		charID := reader.ReadInt32()

		channel.Players.OnCharacterFromID(charID, func(recipient *channel.MapleCharacter) {
			channel.ActiveRooms.OnConn(conn, func(r *channel.Room) {
				for _, p := range r.Participants {
					if p == nil {
						// send request trade packet
						channel.Players.OnCharacterFromConn(conn, func(sender *channel.MapleCharacter) {
							recipient.SendPacket(packets.RoomTradeInvite(sender.GetName(), r.ID))
						})
					}
				}
			})
		})
	}
}
