package server

import (
	"fmt"
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
	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	srcInst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	switch entryType {
	case 0:
		if char.HP() == 0 {
			returnMapID := field.Data.ReturnMap
			portal, err := srcInst.GetRandomSpawnPortal()

			if err == nil {
				conn.Send(entity.PacketPlayerNoChange())
				return
			}

			server.WarpPlayer(player, returnMapID, portal.ID())
			player.SetHP(50)
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := srcInst.GetPortalFromName(portalName)

		if !player.CheckPos(srcPortal.Pos(), 60, 10) { // I'm guessing what the portal hit box is
			if conn.GetAdminLevel() > 0 {
				conn.Send(entity.PacketMessageRedText("Portal - " + srcPortal.Pos().String() + " Player - " + player.Pos().String()))
			}

			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		dstField, ok := server.fields[srcPortal.DestFieldID()]

		if !ok {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		dstInst, err := dstField.GetInstance(player.InstanceID())

		if err != nil {
			if dstInst, err = dstField.GetInstance(0); err != nil {
				return
			}
		}

		dstPortal, err := dstInst.GetPortalFromName(srcPortal.DestName())

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		server.WarpPlayer(player, dstField.ID, dstPortal.ID())

	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
}

func (server* ChannelServer) WarpPlayer(player *entity.Player, mapID int32, mapPos byte) error {
	srcField, ok := server.fields[player.Char().MapID()]

	if !ok {
		return fmt.Errorf("Error in map id %d", player.Char().MapID())
	}

	srcInst, err := srcField.GetInstance(player.InstanceID())

	if err != nil {
		return err
	}

	dstField, ok := server.fields[mapID]

	if !ok {
		return fmt.Errorf("Error in map id %d", mapID)
	}

	dstInst, err := dstField.GetInstance(player.InstanceID())

	if err != nil {
		if dstInst, err = dstField.GetInstance(0); err != nil { // Check player is not in higher level instance than available
			return err
		}

		player.SetInstance(0)
	}

	dstPortal, err := dstInst.GetPortalFromID(mapPos)

	if err != nil {
		return fmt.Errorf("Error in portal id %d", mapPos)
	}

	srcInst.RemovePlayer(player)

	player.SetMapID(dstField.ID)
	player.SetMapPosID(dstPortal.ID())
	player.SetPos(dstPortal.Pos())
	player.SetFoothold(0) // Why is this needed to prevent incorrect initial x,y for others?
	player.Send(entity.PacketMapChange(dstField.ID, int32(server.id), dstPortal.ID(), player.Char().HP()))

	dstInst.AddPlayer(player)

	return nil
}

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating = append(server.migrating, conn)
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

	char := player.Char()
	char.Save(server.db)

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(entity.PacketMessageDialogueBox("Cannot change channel"))
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, char.ID())

			if err != nil {
				log.Println(err)
				return
			}

			conn.Send(entity.PacketChangeChannel(server.channels[id].ip, server.channels[id].port))
		}
	}
}

func (server ChannelServer) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

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

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerMove(char.ID(), moveBytes), conn)
}

func (server *ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, _ := server.players.GetFromConn(conn)
	char := player.Char()

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerEmoticon(char.ID(), emote), conn)
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
		return
	}

	conn.SetAccountID(accountID)
	char := entity.Character{}
	char.LoadFromID(server.db, charID)

	var adminLevel int
	err = server.db.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAdminLevel(adminLevel)

	_, err = server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", -1, charID)

	if err != nil {
		log.Println(err)
		return
	}

	server.players = append(server.players, entity.NewPlayer(conn, char))

	conn.Send(entity.PacketPlayerEnterGame(char, int32(server.id)))
	conn.Send(entity.PacketMessageScrollingHeader("Valhalla Archival Project"))

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(0)

	if err != nil {
		return
	}

	inst.AddPlayer(server.players[len(server.players)-1])
}
