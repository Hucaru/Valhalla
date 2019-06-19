package game

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating[conn] = id
	server.players[conn].char.save(server.db)

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(packetMessageDialogueBox("Cannot change channel"))
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, server.players[conn].char.id)

			if err != nil {
				panic(err)
			}

			conn.Send(packetChangeChannel(server.channels[id].ip, server.channels[id].port))
		}
	}
}

func (server *ChannelServer) playerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var migrationID byte
	err := server.db.QueryRow("SELECT migrationID FROM characters WHERE id=?", charID).Scan(&migrationID)

	if err != nil {
		log.Println(err)
		return
	}

	if migrationID != server.id {
		return
	}

	var accountID int32
	err = server.db.QueryRow("SELECT accountID FROM characters WHERE id=?", charID).Scan(&accountID)

	if err != nil {
		log.Println(err)
	}

	conn.SetAccountID(accountID)

	// check migration

	char := character{}
	char.loadFromID(server.db, charID)

	var adminLevel int
	err = server.db.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
	}

	conn.SetAdminLevel(adminLevel)

	_, err = server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", -1, charID)

	if err != nil {
		panic(err)
	}

	server.players[conn] = newPlayer(conn, char)

	conn.Send(packetPlayerEnterGame(char, int32(server.id)))
	conn.Send(packetMessageScrollingHeader("Valhalla Archival Project"))

	server.fields[char.mapID].addPlayer(conn, server.players[conn].instanceID)
}

func (server *ChannelServer) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	player := server.players[conn]
	char := player.char

	if char.portalCount != reader.ReadByte() {
		return
	}

	moveData, finalData := parseMovement(reader)

	if !moveData.validateChar(char) {
		return
	}

	moveBytes := generateMovementBytes(moveData)

	player.updateMovement(finalData)

	server.fields[char.mapID].sendExcept(packetPlayerMove(char.id, moveBytes), conn, player.instanceID)
}

func (server *ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player := server.players[conn]
	char := player.char

	server.fields[char.mapID].sendExcept(packetPlayerEmoticon(char.id, emote), conn, player.instanceID)
}
