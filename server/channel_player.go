package server

import (
	"log"

	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) playerUsePortal(conn mnet.Client, reader mpacket.Reader) {
	player, _ := server.players.GetFromConn(conn)
	char := player.Char()

	if char.PortalCount() != reader.ReadByte() {
		conn.Send(entity.PacketPlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()
	field := server.fields[char.MapID()]

	switch entryType {
	case 0:
		if char.HP() == 0 {
			returnMapID := field.Data.ReturnMap
			portal, err := field.GetRandomSpawnPortal()

			if err == nil {
				conn.Send(entity.PacketPlayerNoChange())
				return
			}

			field.RemovePlayer(conn, player.InstanceID())
			player.SetMapID(returnMapID)
			player.SetPos(portal.Pos())
			conn.Send(entity.PacketMapChange(returnMapID, int32(server.id), portal.ID(), char.HP()))
			server.fields[returnMapID].AddPlayer(conn, player.InstanceID())

			player.SetHP(50)
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := field.GetPortalFromName(portalName)

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		dstField := server.fields[srcPortal.DestFieldID()]
		dstPortal, err := dstField.GetPortalFromName(srcPortal.DestName())

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		field.RemovePlayer(conn, player.InstanceID())
		player.SetMapID(dstField.ID)
		player.SetPos(dstPortal.Pos())
		conn.Send(entity.PacketMapChange(dstField.ID, int32(server.id), dstPortal.ID(), char.HP()))
		server.fields[dstField.ID].AddPlayer(conn, player.InstanceID())

	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
}

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating = append(server.migrating, conn)
	player, _ := server.players.GetFromConn(conn)
	char := player.Char()
	char.Save(server.db)

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(entity.PacketMessageDialogueBox("Cannot change channel"))
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, char.ID())

			if err != nil {
				panic(err)
			}

			conn.Send(entity.PacketChangeChannel(server.channels[id].ip, server.channels[id].port))
		}
	}
}

func (server *ChannelServer) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	player, _ := server.players.GetFromConn(conn)
	char := player.Char()

	if char.PortalCount() != reader.ReadByte() {
		return
	}

	moveData, finalData := entity.ParseMovement(reader)

	if !moveData.ValidateChar(char) {
		return
	}

	moveBytes := entity.GenerateMovementBytes(moveData)

	player.UpdateMovement(finalData)

	server.fields[char.MapID()].SendExcept(entity.PacketPlayerMove(char.ID(), moveBytes), conn, player.InstanceID())
}

func (server *ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, _ := server.players.GetFromConn(conn)
	char := player.Char()

	server.fields[char.MapID()].SendExcept(entity.PacketPlayerEmoticon(char.ID(), emote), conn, player.InstanceID())
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
	char := entity.Character{}
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

	server.players = append(server.players, entity.NewPlayer(conn, char))

	conn.Send(entity.PacketPlayerEnterGame(char, int32(server.id)))
	conn.Send(entity.PacketMessageScrollingHeader("Valhalla Archival Project"))

	server.fields[char.MapID()].AddPlayer(conn, server.players[len(server.players)-1].InstanceID())
}
