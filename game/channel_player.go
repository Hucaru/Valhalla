package game

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating[conn] = id
	server.sessions[conn].Save(server.db)

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(packetMessageDialogueBox("Cannot change channel"))
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, server.sessions[conn].id)

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
	char.LoadFromID(server.db, charID)

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

	server.sessions[conn] = &char

	conn.Send(packetPlayerEnterGame(char, int32(server.id)))
	conn.Send(packetMessageScrollingHeader("Valhalla Archival Project"))

	server.fields[char.mapID].addPlayer(conn, char.instanceID)
}
