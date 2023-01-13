package channel

import (
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"log"
	rand2 "math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/common/translates"

	"github.com/Hucaru/Valhalla/common/db"
	"github.com/Hucaru/Valhalla/common/proto"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	proto2 "google.golang.org/protobuf/proto"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// HandleClientPacket data
func (server *Server) HandleClientPacket(
	conn *mnet.Client, reader mpacket.Reader, msgProtocolType uint32) {
	server.playerAction(conn, RequestedParam{Num: msgProtocolType, Reader: reader})
}

func (server *Server) playerAction(conn *mnet.Client, reader RequestedParam) {
	if reader.Num == constant.OnConnected {
		c := make(chan RequestedParam, 4096*4)
		server.playerActions.Set(conn.String(), c)
		go func(server *Server, conn *mnet.Client, c chan RequestedParam) {
			for {
				// Kioni
				select {
				case p := <-c:
					if _, ok := server.PlayerActionHandler[p.Num]; ok {
						server.PlayerActionHandler[p.Num](conn, p.Reader)
						if p.Num == constant.OnDisconnected {
							log.Println("constant.OnDisconnected")
							close(c)
							return
						}
					}
				default:
					//log.Println("state : ", runtime.NumGoroutine(), runtime.NumCPU())
					time.Sleep(10 * time.Millisecond)
					runtime.Gosched()
				}
			}
		}(server, conn, c)
		c <- reader
	} else {
		c, ok := server.playerActions.Get(conn.String())
		if ok {
			c <- reader
		}
	}
}

func (server *Server) playerConnect(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestLoginUser{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.UuId) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}
	var player *model.Player

	if msg.IsBot == 1 {
		player, err = db.GetLoggedDataForBot(msg.GetUuId())
	} else {
		player, err = db.GetLoggedData(msg.GetUuId())
		if err != nil {
			db.AddNewAccount(player)
		} else {
			err1 := db.UpdateLoginState(msg.GetUuId(), true)
			if err1 != nil {
				log.Println("Unable to complete login for ", msg.GetUuId())
				m, err2 := proto.ErrorLoginResponse(err.Error(), msg.GetUuId())
				if err2 != nil {
					log.Println("ErrorLoginResponse", err2)
				}
				conn.BaseConn.Send(m)
				return
			}
		}
	}

	if msg.GetSpawnPosX() != 0 {
		player.Character.PosX = msg.GetSpawnPosX()
	}

	if msg.GetSpawnPosY() != 0 {
		player.Character.PosY = msg.GetSpawnPosY()
	}

	if msg.GetSpawnPosZ() != 0 {
		player.Character.PosZ = msg.GetSpawnPosZ()
	}

	player.IsBot = msg.IsBot

	// TMP part, will be moved later
	if player.RegionID == constant.MetaClassRoom {
		player.RegionID = constant.MetaSchool
		player.Character.PosX = -8597
		player.Character.PosY = -23392
		player.Character.PosZ = 2180

		db.UpdateRegionID(player.CharacterID, int32(player.RegionID))
	}

	//plr := loadPlayer(conn, *msg)
	//plr.rates = &server.rates
	//plr.conn.SetPlayer(*player)
	//
	//server.addPlayer(&plr)
	server.clients.Set(msg.UuId, conn)
	conn.SetPlayer(*player)

	if msg.IsBot == 1 {
		conn.GetPlayer().Character.Top = constant.RandomTop[rand2.Intn(4)]
		conn.GetPlayer().Character.Bottom = constant.RandomBottom[rand2.Intn(4)]
		conn.GetPlayer().Character.Clothes = constant.RandomClothes[rand2.Intn(4)]
		conn.GetPlayer().Character.Hair = constant.RandomHair[rand2.Intn(5)]
		//go server.addToEmulateMove(&plr)
		//return
	}

	response := proto.AccountReport(conn.GetPlayer().UId, conn.GetPlayer().Character)
	server.sendMsgToRegion(conn, response, constant.P2C_ReportLoginUser)

	account := proto.AccountResult(player)
	fmt.Println("NumGoroutine COUNT CONNECT", runtime.NumGoroutine())
	x, y := common.FindGrid(player.Character.PosX, player.Character.PosY)
	loggedPlayers := server.getPlayersOnGrids(x, y, conn.GetPlayer().UId)
	if loggedPlayers != nil {
		//log.Println("START MOVING EMULATION")

		server.fMovePlayers = append(server.fMovePlayers, PlayerMovement{
			name: msg.UuId,
			x:    player.Character.PosX,
			y:    player.Character.PosY,
		})

		//go server.addToEmulateMoving(plr.conn.GetPlayer().UId, loggedPlayers)
		fmt.Println("NumGoroutine COUNT EMULATE", runtime.NumGoroutine())
		users := server.convertPlayersToLoginResult(loggedPlayers)
		account.LoggedUsers = append(account.LoggedUsers, users...)
	}

	//fmt.Println(" Client at ", conn, "UID:", msg.GetUuId(), "LOCATION:", player.Character.PosX, player.Character.PosY)

	GridX, GridY := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
	server.gridMgr.Add(GridX, GridY, conn)

	server.sendMsgToMe(conn, account, constant.P2C_ResultLoginUser)
	response = nil
}

func (server *Server) getRoomPlayers(uID string, mX, mY float32) []*model.Player {
	plrs := make([]*model.Player, 0)

	//x, y := common.FindGrid(mX, mY)
	//loggedPlayers := server.getPlayersOnGrids(x, y, uID)
	//
	//for i := 0; i < len(loggedPlayers); i++ {
	//	if uID == loggedPlayers[i].conn.GetPlayer().UId {
	//		continue
	//	}
	//	if loggedPlayers[i].conn.GetPlayer().Interaction != nil {
	//		plrs = append(plrs, loggedPlayers[i].conn.GetPlayer())
	//	}
	//
	//}
	return plrs
}

func (server *Server) setPlayer(plr *model.Player) {
	//SomeMapMutex.Lock()
	//server.players[plr.UId].conn.SetPlayer(*plr)
	//SomeMapMutex.Unlock()
}

func (server *Server) isPlayerOnline(uID string) bool {
	//SomeMapMutex.RLock()
	//_, ok := server.players[uID]
	//SomeMapMutex.RUnlock()
	return server.clients.Has(uID)
}

func (server *Server) sendMsgToMe(conn *mnet.Client, msg proto2.Message, msgType int) {
	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	//plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}
	conn.Send(res)
}

func (server *Server) sendMsgToPlayer(msg proto2.Message, uID string, msgType int) {

	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	c, ok := server.clients.Get(uID)
	if ok {
		(*c).Send(res)
	}
	//SomeMapMutex.RLock()
	//_, ok := server.players[uID]
	//if ok {
	//	server.players[uID].conn.Send(res)
	//}
	//SomeMapMutex.RUnlock()
}

func (server *Server) sendMsgToAll(msg proto2.Message, uID string, msgType int) {
	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	itemChan := server.clients.IterBuffered()

	for item := range itemChan {
		if (*item.Val).GetPlayer().UId == uID {
			continue
		}

		(*item.Val).Send(res)
	}

	//for _, v := range server.clients {
	//	if uID == (*v).GetPlayer().UId {
	//		continue
	//	}
	//	(*v).Send(res)
	//}
}

func (server *Server) sendMsgToRegion(conn *mnet.Client, msg proto2.Message, msgType int) {

	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	log.Println("player not found", err)
	//	return
	//}

	x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
	plrs := server.getPlayersOnGrids(x, y, conn.GetPlayer().UId)

	//log.Println("getPlayersOnGrids", len(plrs))

	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	for _, p := range plrs {
		if conn.GetPlayer().UId == (*p).GetPlayer().UId {
			continue
		}
		(*p).Send(res)
	}
}

func (server *Server) playerChangeChannel(conn *mnet.Client, reader mpacket.Reader) {
	//msg := &mc_metadata.C2P_RequestRegionChange{}
	//err := proto.Unmarshal(reader.GetBuffer(), msg)
	//if err != nil || len(msg.UuId) == 0 {
	//	log.Println("Failed to parse data:", err)
	//	return
	//}
	//
	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return
	//}
	//
	//db.UpdateRegionID(plr.conn.GetPlayer().CharacterID, msg.GetRegionId())
	//
	//responseOld := proto.ChannelChangeForOldReport(plr.conn.GetPlayer().UId, plr.conn.GetPlayer().Character)
	//log.Println("REGION_CHANGED PREV REGION SEND", plr.conn.GetPlayer().RegionID)
	//go server.sendMsgToRegion(conn, responseOld, constant.P2C_ReportRegionLeave)
	//
	//responseNew := proto.ChannelChangeForNewReport(plr.conn.GetPlayer())
	//log.Println("REGION_CHANGED TO ", msg.GetRegionId())
	//go server.sendMsgToRegion(conn, responseNew, constant.P2C_ReportRegionChange)
	//
	//plr.conn.GetPlayer().RegionID = int64(msg.RegionId)
	//server.setPlayer(plr.conn.GetPlayer())
	//
	//account := proto.RegionResult(plr.conn.GetPlayer())
	//
	//x, y := common.FindGrid(plr.conn.GetPlayer().Character.PosX, plr.conn.GetPlayer().Character.PosY)
	//loggedAccounts := server.getPlayersOnGrids(x, y, plr.conn.GetPlayer().UId)
	//
	//if err != nil {
	//	log.Println("ERROR GetLoggedUsersData", plr.conn.GetPlayer().UId)
	//	return
	//}
	//
	//users := server.convertPlayersToRegionReport(loggedAccounts)
	//account.RegionUsers = append(account.RegionUsers, users...)
	//server.sendMsgToMe(conn, account, constant.P2C_ResultRegionChange)
}

func (server *Server) playerMovementStart(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestMoveStart{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	//if server.isCellChanged(conn, msg.GetMovementData()) {
	//	oldPlr, newPlr := server.getNineCellsPlayers(conn, msg.GetMovementData())
	//	cellsData := &mc_metadata.P2C_ResultGrid{
	//		OldPlayers: server.convertPlayersToGridChanged(oldPlr),
	//		NewPlayers: server.convertPlayersToGridChanged(newPlr),
	//	}
	//	//server.removeFromEmulateMoving(conn.GetPlayer().UId, oldPlr)
	//	//server.addToEmulateMoving(conn.GetPlayer().UId, newPlr)
	//
	//	//server.switchPlayerCell(conn, msg.GetMovementData())
	//	go server.sendMsgToMe(conn, cellsData, constant.P2C_ResultGrid)
	//}

	//res := &mc_metadata.P2C_ReportMoveStart{
	//	MovementData: msg.GetMovementData(),
	//}

	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMoveStart)

	//server.sendMsgToRegion(conn, res, constant.P2C_ReportMoveStart)
	//server.updateUserLocation(conn, msg.GetMovementData())
}

func (server *Server) playerMovementEnd(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestMoveEnd{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	//if server.isCellChanged(conn, msg.GetMovementData()) {
	//	oldPlr, newPlr := server.getNineCellsPlayers(conn, msg.GetMovementData())
	//	cellsData := &mc_metadata.P2C_ResultGrid{
	//		OldPlayers: server.convertPlayersToGridChanged(oldPlr),
	//		NewPlayers: server.convertPlayersToGridChanged(newPlr),
	//	}
	//	//server.removeFromEmulateMoving(conn.GetPlayer().UId, oldPlr)
	//	//server.addToEmulateMoving(conn.GetPlayer().UId, newPlr)
	//
	//	//server.switchPlayerCell(conn, msg.GetMovementData())
	//	go server.sendMsgToMe(conn, cellsData, constant.P2C_ResultGrid)
	//}

	//res := &mc_metadata.P2C_ReportMoveEnd{
	//	MovementData: msg.GetMovementData(),
	//}

	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMoveEnd)

	//server.sendMsgToRegion(conn, res, constant.P2C_ReportMoveEnd)
	//server.updateUserLocation(conn, msg.GetMovementData())
}

func (server *Server) getPlayersOnGrids(x, y int, uID string) map[string]*mnet.Client {
	oldList := map[string]*mnet.Client{}

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			oldGridX := x + i
			oldGridY := y + j

			maps.Copy(oldList, server.gridMgr.FillPlayers(oldGridX, oldGridY))
		}
	}

	delete(oldList, uID)

	return oldList

	//for _, v := range oldList {
	//	if (*v).GetPlayer().UId == uID {
	//		continue
	//	}
	//
	//	//p, err := server.players.getFromConn(*v)
	//	//
	//	//if err != nil {
	//	//	log.Println("getPlayersOnGrids :", err)
	//	//	continue
	//	//}
	//
	//	(*v).send
	//
	//	players = append(players, p)
	//}

	// return players

	////main
	//SomeMapMutex.RLock()
	//arr := server.getGridPlayers(x, y)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////left-top
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x-1, y+1)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////top
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x, y+1)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////right-top
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x+1, y+1)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////right
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x+1, y)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////right-bottom
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x+1, y-1)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////bottom
	//arr = server.getGridPlayers(x, y-1)
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////left-bottom
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x-1, y-1)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//
	////left
	//SomeMapMutex.RLock()
	//arr = server.getGridPlayers(x-1, y)
	//SomeMapMutex.RUnlock()
	//for _, val := range arr {
	//	if val.conn.GetPlayer().UId == uID {
	//		continue
	//	}
	//	players = append(players, val)
	//}
	//arr = nil
	//return players

}

func (server *Server) existsPlayerFromGrid(uID string, x, y int) bool {
	plrs := server.getGridPlayers(x, y)
	for i := 0; i < len(plrs); i++ {
		return plrs[i] != nil && plrs[i].conn.GetPlayer().UId == uID
	}
	return false
}

func (server *Server) getGridPlayers(x int, y int) map[int]*player {
	if len(server.mapGrid) <= x {
		return map[int]*player{}
	}
	if x < 0 {
		x = 1
	}

	if len(server.mapGrid[x]) <= y {
		return map[int]*player{}
	}

	if y < 0 {
		y = 1
	}

	return server.mapGrid[x][y]
}

func (server *Server) playerInteraction(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestInteractionAttach{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	errR := errors.New("error")
	errR = nil

	log.Println("PreError: ObjIndex:", msg.GetObjectIndex())
	log.Println("PreError: Attach:", msg.GetAttachEnable())

	if msg.GetAttachEnable() == 0 {
		errR = server.InsertInteractionAndSend(conn, msg)
		if errR != nil {
			log.Println("Error: Insert:", errR)
			return
		}

	} else {
		errR = server.DeleteInteractionAndSend(conn, msg)
		if errR != nil {
			log.Println("Error: Delete:", errR)
			return
		}
	}

	res := &mc_metadata.P2C_ReportInteractionAttach{
		UuId:            msg.GetUuId(),
		AttachEnable:    msg.GetAttachEnable(),
		ObjectIndex:     msg.GetObjectIndex(),
		AnimMontageName: msg.GetAnimMontageName(),
		DestinationX:    msg.GetDestinationX(),
		DestinationY:    msg.GetDestinationY(),
		DestinationZ:    msg.GetDestinationZ(),
	}
	log.Println("P2C_ReportInteractionAttach sent from ", res.GetUuId())
	server.sendMsgToAll(res, msg.GetUuId(), constant.P2C_ReportInteractionAttach)
}

func (server *Server) DeleteInteractionAndSend(conn *mnet.Client, msg *mc_metadata.C2P_RequestInteractionAttach) error {

	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return err
	//}
	//
	//att := &mc_metadata.P2C_ResultInteractionAttach{
	//	ErrorCode: -1,
	//}
	//
	//if plr.conn.GetPlayer().Interaction == nil {
	//	plr.conn.GetPlayer().Interaction = &model.Interaction{}
	//}
	//plr.conn.GetPlayer().Interaction.ObjectIndex = msg.GetObjectIndex()
	//plr.conn.GetPlayer().Interaction.AttachEnabled = msg.GetAttachEnable()
	//plr.conn.GetPlayer().Interaction.AnimMontageName = msg.GetAnimMontageName()
	//plr.conn.GetPlayer().Interaction.DestinationX = msg.GetDestinationX()
	//plr.conn.GetPlayer().Interaction.DestinationY = msg.GetDestinationY()
	//plr.conn.GetPlayer().Interaction.DestinationZ = msg.GetDestinationZ()
	//server.setPlayer(plr.conn.GetPlayer())
	//
	//server.sendMsgToMe(conn, att, constant.P2C_ResultInteractionAttach)
	return nil
}

func (server *Server) InsertInteractionAndSend(conn *mnet.Client, msg *mc_metadata.C2P_RequestInteractionAttach) error {

	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return nil
	//}
	//
	//att := &mc_metadata.P2C_ResultInteractionAttach{
	//	ErrorCode: -1,
	//}
	//
	//x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosX)
	//users := server.getGridPlayers(x, y)
	//
	//for i := 0; i < len(users); i++ {
	//	if users[i].conn.GetPlayer().Interaction != nil &&
	//		plr.conn.GetPlayer().UId != users[i].conn.GetPlayer().UId &&
	//		msg.ObjectIndex == users[i].conn.GetPlayer().Interaction.ObjectIndex {
	//		att.ErrorCode = constant.ErrorCodeChairNotEmpty
	//		break
	//	}
	//}
	//
	//if att.ErrorCode == -1 {
	//	if plr.conn.GetPlayer().Interaction == nil {
	//		plr.conn.GetPlayer().Interaction = &model.Interaction{}
	//	}
	//	plr.conn.GetPlayer().Interaction.ObjectIndex = msg.GetObjectIndex()
	//	plr.conn.GetPlayer().Interaction.AttachEnabled = msg.GetAttachEnable()
	//	plr.conn.GetPlayer().Interaction.AnimMontageName = msg.GetAnimMontageName()
	//	plr.conn.GetPlayer().Interaction.DestinationX = msg.GetDestinationX()
	//	plr.conn.GetPlayer().Interaction.DestinationY = msg.GetDestinationY()
	//	plr.conn.GetPlayer().Interaction.DestinationZ = msg.GetDestinationZ()
	//	server.setPlayer(plr.conn.GetPlayer())
	//}
	//
	//server.sendMsgToMe(conn, att, constant.P2C_ResultInteractionAttach)
	//if att.ErrorCode == -1 {
	//	return nil
	//} else {
	//	return errors.New("chair not empty")
	//}

	return nil
}

func (server *Server) playerPlayAnimation(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestPlayMontage{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	res := &mc_metadata.P2C_ReportPlayMontage{
		UuId:    msg.GetUuId(),
		AnimTid: msg.GetAnimTid(),
	}

	go server.sendMsgToRegion(conn, res, constant.P2C_ReportPlayMontage)
}

func (server *Server) playerRegionRoleChecking(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestRoleChecking{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	x, y := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosX)
	users := server.getGridPlayers(x, y)

	is := 0
	for i := 0; i < len(users); i++ {
		if users[i].conn.GetPlayer().Interaction.IsInteraction &&
			users[i].conn.GetPlayer().Character.Role > 1 {
			is = 1
			break
		}
	}

	res := &mc_metadata.P2C_ResultRoleChecking{
		UuId:      msg.GetUuId(),
		IsTeacher: int32(is),
	}

	log.Println("P2C_ResultRoleChecking")
	server.sendMsgToMe(conn, res, constant.P2C_ResultRoleChecking)
}

func (server *Server) playerEnterToRoom(conn *mnet.Client, reader mpacket.Reader) {

	//msg := &mc_metadata.C2P_RequestMetaSchoolEnter{}
	//err := proto.Unmarshal(reader.GetBuffer(), msg)
	//if err != nil {
	//	log.Println("Failed to parse data:", err)
	//	return
	//}
	//
	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return
	//}
	//
	//plr.conn.GetPlayer().Character.Role = msg.TeacherEnable
	//plr.conn.GetPlayer().Interaction = &model.Interaction{
	//	AttachEnabled: 1,
	//	ObjectIndex:   -1,
	//}
	//server.setPlayer(plr.conn.GetPlayer())
	//
	//reportEnter := &mc_metadata.P2C_ReportMetaSchoolEnter{
	//	UuId:          msg.GetUuId(),
	//	TeacherEnable: msg.GetTeacherEnable(),
	//	PlayerInfo: &mc_metadata.P2C_PlayerInfo{
	//		UuId:     plr.conn.GetPlayer().UId,
	//		Nickname: plr.conn.GetPlayer().Character.NickName,
	//		Role:     plr.conn.GetPlayer().Character.Role,
	//		Hair:     plr.conn.GetPlayer().Character.Hair,
	//		Top:      plr.conn.GetPlayer().Character.Top,
	//		Bottom:   plr.conn.GetPlayer().Character.Bottom,
	//		Clothes:  plr.conn.GetPlayer().Character.Clothes,
	//	},
	//}
	//
	//log.Println("P2C_ResultMetaSchoolEnter sendMsgToRegion")
	//go server.sendMsgToRegion(conn, reportEnter, constant.P2C_ReportMetaSchoolEnter)
	//
	//res := &mc_metadata.P2C_ResultMetaSchoolEnter{
	//	UuId:          msg.GetUuId(),
	//	TeacherEnable: msg.GetTeacherEnable(),
	//	DataSchool:    proto.ConvertPlayersToRoomReport(server.getRoomPlayers(msg.GetUuId(), plr.conn.GetPlayer().Character.PosX, plr.conn.GetPlayer().Character.PosY)),
	//}
	//
	//server.sendMsgToMe(conn, res, constant.P2C_ResultMetaSchoolEnter)
}

func (server *Server) playerLeaveFromRoom(conn *mnet.Client, reader mpacket.Reader) {

	//msg := &mc_metadata.C2P_RequestMetaSchoolLeave{}
	//err := proto.Unmarshal(reader.GetBuffer(), msg)
	//if err != nil {
	//	log.Println("Failed to parse data:", err)
	//	return
	//}
	//
	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	return
	//}
	//
	//plr.conn.GetPlayer().Interaction = nil
	//server.setPlayer(plr.conn.GetPlayer())
	//
	//res := &mc_metadata.P2C_ReportMetaSchoolLeave{
	//	UuId: msg.GetUuId(),
	//	PlayerInfo: &mc_metadata.P2C_PlayerInfo{
	//		UuId:     plr.conn.GetPlayer().UId,
	//		Nickname: plr.conn.GetPlayer().Character.NickName,
	//		Role:     plr.conn.GetPlayer().Character.Role,
	//		Hair:     plr.conn.GetPlayer().Character.Hair,
	//		Top:      plr.conn.GetPlayer().Character.Top,
	//		Bottom:   plr.conn.GetPlayer().Character.Bottom,
	//		Clothes:  plr.conn.GetPlayer().Character.Clothes,
	//	},
	//}
	//go server.sendMsgToRegion(conn, res, constant.P2C_ReportMetaSchoolLeave)
}

func (server *Server) playerMovement(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestMove{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMove)

	//if server.isCellChanged(conn, msg.GetMovementData()) {
	//	oldPlr, newPlr := server.getNineCellsPlayers(conn, msg.GetMovementData())
	//	if len(oldPlr) > 0 || len(newPlr) > 0 {
	//		cellsData := &mc_metadata.P2C_ResultGrid{
	//			OldPlayers: server.convertPlayersToGridChanged(oldPlr),
	//			NewPlayers: server.convertPlayersToGridChanged(newPlr),
	//		}
	//		server.sendMsgToMe(conn, cellsData, constant.P2C_ResultGrid)
	//	}
	//	server.switchPlayerCell(conn, msg.GetMovementData())
	//	fmt.Println("NumGoroutine switchPlayerCell", runtime.NumGoroutine())
	//	//go server.addToEmulateMoving(conn.GetPlayer().UId, newPlr)
	//
	//	for i := 0; i < len(oldPlr); i++ {
	//		go server.sendMsgToPlayer(&mc_metadata.P2C_ReportGridOld{
	//			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
	//				UuId: msg.GetMovementData().UuId,
	//			},
	//		}, oldPlr[i].conn.GetPlayer().UId, constant.P2C_ReportGridOld)
	//	}
	//
	//	for i := 0; i < len(newPlr); i++ {
	//		go server.sendMsgToPlayer(&mc_metadata.P2C_ReportGridNew{
	//			SpawnPosX: msg.GetMovementData().DestinationX,
	//			SpawnPosY: msg.GetMovementData().DestinationY,
	//			SpawnPosZ: msg.GetMovementData().DestinationZ,
	//			SpawnRotX: msg.GetMovementData().DeatinationRotationX,
	//			SpawnRotY: msg.GetMovementData().DeatinationRotationY,
	//			SpawnRotZ: msg.GetMovementData().DeatinationRotationZ,
	//			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
	//				Nickname: msg.GetMovementData().UuId,
	//				UuId:     msg.GetMovementData().UuId,
	//				Top:      conn.GetPlayer().Character.Top,
	//				Bottom:   conn.GetPlayer().Character.Bottom,
	//				Clothes:  conn.GetPlayer().Character.Clothes,
	//				Hair:     conn.GetPlayer().Character.Hair,
	//			},
	//		}, newPlr[i].conn.GetPlayer().UId, constant.P2C_ReportGridNew)
	//	}
	//
	//	fmt.Println("NumGoroutine addToEmulateMoving", runtime.NumGoroutine())
	//}

	//res := &mc_metadata.P2C_ReportMove{
	//	MovementData: msg.GetMovementData(),
	//}
	//
	//server.sendMsgToRegion(conn, res, constant.P2C_ReportMove)
	//server.updateUserLocation(conn, msg.GetMovementData())

}

func (server *Server) moveProcess(conn *mnet.Client, x, y float32, uId string, movement *mc_metadata.Movement, moveType int) {
	addList, removeList, aroundList := server.gridMgr.OnMove(x, y, uId)

	for k, v := range addList {
		res := &mc_metadata.P2C_ReportGridNew{
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: k,
				UuId:     k,
				Top:      (*v).GetPlayer().Character.Top,
				Bottom:   (*v).GetPlayer().Character.Bottom,
				Clothes:  (*v).GetPlayer().Character.Clothes,
				Hair:     (*v).GetPlayer().Character.Hair,
			},
			SpawnPosX: (*v).GetPlayer().Character.PosX,
			SpawnPosY: (*v).GetPlayer().Character.PosY,
			SpawnPosZ: (*v).GetPlayer().Character.PosZ,
			SpawnRotX: (*v).GetPlayer().Character.RotX,
			SpawnRotY: (*v).GetPlayer().Character.RotY,
			SpawnRotZ: (*v).GetPlayer().Character.RotZ,
		}

		res2 := &mc_metadata.P2C_ReportGridNew{
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId:     conn.GetPlayer().UId,
				Nickname: conn.GetPlayer().UId,
				Top:      conn.GetPlayer().Character.Top,
				Bottom:   conn.GetPlayer().Character.Bottom,
				Clothes:  conn.GetPlayer().Character.Clothes,
				Hair:     conn.GetPlayer().Character.Hair,
			},
			SpawnPosX: conn.GetPlayer().Character.PosX,
			SpawnPosY: conn.GetPlayer().Character.PosY,
			SpawnPosZ: conn.GetPlayer().Character.PosZ,
			SpawnRotX: conn.GetPlayer().Character.RotX,
			SpawnRotY: conn.GetPlayer().Character.RotY,
			SpawnRotZ: conn.GetPlayer().Character.RotZ,
		}

		server.sendMsgToMe(conn, res, constant.P2C_ReportGridNew)
		server.sendMsgToMe(v, res2, constant.P2C_ReportGridNew)
	}

	for k, v := range removeList {
		res := &mc_metadata.P2C_ReportGridOld{
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId: k,
			},
		}

		res2 := &mc_metadata.P2C_ReportGridOld{
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId: conn.GetPlayer().UId,
			},
		}

		//fmt.Println(fmt.Sprintf("conn : %s v : %s res : %s res2 : %s", conn.GetPlayer().UId, (*v).GetPlayer().UId, res.PlayerInfo.UuId, res2.PlayerInfo.UuId))

		server.sendMsgToMe(conn, res, constant.P2C_ReportGridOld)
		server.sendMsgToMe(v, res2, constant.P2C_ReportGridOld)
	}

	switch moveType {
	case constant.P2C_ReportMoveStart:
		res := &mc_metadata.P2C_ReportMoveStart{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, res, constant.P2C_ReportMoveStart)
		}
	case constant.P2C_ReportMove:
		res := &mc_metadata.P2C_ReportMove{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, res, constant.P2C_ReportMove)
		}
	case constant.P2C_ReportMoveEnd:
		res := &mc_metadata.P2C_ReportMoveEnd{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, res, constant.P2C_ReportMoveEnd)
		}
	}

	conn.GetPlayer().Character.PosX = movement.GetDestinationX()
	conn.GetPlayer().Character.PosY = movement.GetDestinationY()
	conn.GetPlayer().Character.PosZ = movement.GetDestinationZ()
	conn.GetPlayer().Character.RotX = movement.GetDeatinationRotationX()
	conn.GetPlayer().Character.RotY = movement.GetDeatinationRotationY()
	conn.GetPlayer().Character.RotZ = movement.GetDeatinationRotationZ()
}

func (server *Server) playerInfo(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestPlayerInfo{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := &mc_metadata.P2C_ResultPlayerInfo{
		ErrorCode: constant.NoError,
	}

	if server.isPlayerOnline(msg.GetUuId()) {
		res.ErrorCode = constant.ErrorCodeAlreadyOnline

		data, err := proto.MakeResponse(res, constant.P2C_ResultPlayerInfo)
		if err != nil {
			log.Println("ERROR P2C_ResultPlayerInfo Already Online", msg.GetUuId())
			return
		}
		conn.Send(data)
		return
	}

	plr, err1 := db.GetLoggedDataByName(msg)

	if err1 != nil {
		//log.Println("Inserting new user playerInfo", fmt.Sprintf("niickname=%s uid=%s", msg.GetNickname(), msg.GetUuId()))
		iErr := db.AddNewAccount(plr)
		if iErr != nil {
			res.ErrorCode = constant.ErrorCodeDuplicateUID
		}
	} else {
		db.UpdatePlayerInfo(
			plr.CharacterID,
			msg.GetHair(),
			msg.GetTop(),
			msg.GetBottom(),
			msg.GetClothes())
	}

	data, err := proto.MakeResponse(res, constant.P2C_ResultPlayerInfo)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", msg.GetUuId())
		return
	}

	conn.BaseConn.Send(data)

	//server.sendMsgToMe(data, conn)
	data = nil
}

func (server *Server) playerLogout(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestLogoutUser{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := &mc_metadata.P2C_ReportLogoutUser{
		UuId: msg.GetUuId(),
	}

	server.sendMsgToAll(res, msg.GetUuId(), constant.P2C_ReportLogoutUser)
	server.removePlayer(conn)
}

func (server *Server) chatSendAll(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestAllChat{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := &mc_metadata.P2C_ReportAllChat{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		Translate: &mc_metadata.P2C_Translate{},
	}

	papagoTranslate := server.translateMessage(msg.GetChat())
	if papagoTranslate != nil {
		res.Translate = papagoTranslate
	}

	server.sendMsgToAll(res, msg.GetUuId(), constant.P2C_ReportAllChat)
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	db.AddPublicMessage(plr.conn.GetPlayer().CharacterID, constant.World, msg.GetChat())

	toMe := &mc_metadata.P2C_ResultAllChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	server.sendMsgToMe(conn, toMe, constant.P2C_ResultAllChat)
}

func (server *Server) chatSendRegion(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestRegionChat{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := &mc_metadata.P2C_ReportRegionChat{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		Translate: &mc_metadata.P2C_Translate{},
	}

	papagoTranslate := server.translateMessage(msg.GetChat())
	if papagoTranslate != nil {
		res.Translate = papagoTranslate
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	go server.sendMsgToRegion(conn, res, constant.P2C_ReportRegionChat)
	db.AddPublicMessage(
		plr.conn.GetPlayer().CharacterID,
		plr.conn.GetPlayer().RegionID,
		msg.GetChat())

	toMe := &mc_metadata.P2C_ResultRegionChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	server.sendMsgToMe(conn, toMe, constant.P2C_ResultRegionChat)
}

func (server *Server) chatSendWhisper(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestWhisper{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	//plr, err := server.players.getFromConn(conn)
	//if err != nil {
	//	log.Println("player not found", err)
	//	return
	//}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := &mc_metadata.P2C_ReportWhisper{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetPlayerNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		Translate: &mc_metadata.P2C_Translate{},
	}

	papagoTranslate := server.translateMessage(msg.GetChat())
	if papagoTranslate != nil {
		res.Translate = papagoTranslate
	}

	toMe := &mc_metadata.P2C_ResultWhisper{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetTargetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		ErrorCode: constant.NoError,
	}

	targetPlayer := server.findCharacterIDByNickname(msg.GetTargetNickname())
	db.AddWhisperMessage(
		conn.GetPlayer().CharacterID,
		targetPlayer.CharacterID,
		msg.GetChat())

	if targetPlayer != nil {
		server.sendMsgToPlayer(res, targetPlayer.UId, constant.P2C_ReportWhisper)
	} else {
		toMe.ErrorCode = constant.ErrorUserOffline
	}

	server.sendMsgToMe(conn, toMe, constant.P2C_ResultWhisper)
}

func (server *Server) findCharacterIDByUID(uID string) *model.Player {
	//SomeMapMutex.RLock()
	//p, ok := server.players[uID]
	//SomeMapMutex.RUnlock()
	//if ok {
	//	return p.conn.GetPlayer()
	//}
	return nil
}

func (server *Server) findCharacterIDByNickname(nickname string) *model.Player {

	//for _, v := range server.players {
	//	if nickname == v.conn.GetPlayer().Character.NickName {
	//		return v.conn.GetPlayer()
	//	}
	//}
	return nil
}

func (server *Server) isCellChanged(conn *mnet.Client, msg *mc_metadata.Movement) bool {
	x1, y1 := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
	x2, y2 := common.FindGrid(msg.GetDestinationX(), msg.GetDestinationY())

	return x1 != x2 || y1 != y2
}

func (server *Server) switchPlayerCell(conn *mnet.Client, msg *mc_metadata.Movement) {
	// if conn.GetPlayer().ModifiedAt >= (msg.ModifiedAt / 1000) {
	// 	return
	// }

	// conn.GetPlayer().ModifiedAt = msg.ModifiedAt / 1000
	x1, y1 := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
	x2, y2 := common.FindGrid(msg.DestinationX, msg.DestinationY)

	if (x1 == x2 && y1 == y2) && !server.existsPlayerFromGrid(conn.GetPlayer().UId, x1, y1) {
		return
	}

	server.removePlayerFromGrid(
		server.getGridPlayers(x1, y1),
		conn.GetPlayer().UId,
		conn.GetPlayer().Character.PosX,
		conn.GetPlayer().Character.PosY)

	SomeMapMutex.RLock()
	_, ok := server.players[conn.GetPlayer().UId]
	SomeMapMutex.RUnlock()
	if ok {
		server.addPlayerToGrid(server.players[conn.GetPlayer().UId], msg.GetDestinationX(), msg.GetDestinationY())
	}
}

//
//func (server *Server) getNineCellsPlayers(conn *mnet.Client, msg *mc_metadata.Movement) (oldPlr []*player, newPlr []*player) {
//	x1, y1 := common.FindGrid(conn.GetPlayer().Character.PosX, conn.GetPlayer().Character.PosY)
//	x2, y2 := common.FindGrid(msg.GetDestinationX(), msg.GetDestinationY())
//
//	oldPlayers := server.getPlayersOnGrids(x1, y1, conn.GetPlayer().UId)
//	newPlayers := server.getPlayersOnGrids(x2, y2, conn.GetPlayer().UId)
//
//	is := true
//
//	for _, n := range newPlayers {
//		is = true
//	old:
//		for _, o := range oldPlayers {
//			if n. == o.conn {
//				is = false
//				break old
//			}
//		}
//		if is {
//			newPlr = append(newPlr, n)
//		}
//	}
//
//	for _, o := range oldPlayers {
//		is = true
//	newp:
//		for _, n := range newPlayers {
//			if n.conn == o.conn {
//				is = false
//				break newp
//			}
//		}
//		if is {
//			oldPlr = append(oldPlr, o)
//		}
//	}
//	return oldPlr, newPlr
//}

func (server *Server) updateUserLocation(conn *mnet.Client, msg *mc_metadata.Movement) {
	SomeMapMutex.RLock()
	_, ok := server.players[conn.GetPlayer().UId]
	if ok {
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.PosX = msg.GetDestinationX()
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.PosY = msg.GetDestinationY()
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.PosZ = msg.GetDestinationZ()
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.RotX = msg.GetDeatinationRotationX()
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.RotY = msg.GetDeatinationRotationY()
		server.players[conn.GetPlayer().UId].conn.GetPlayer().Character.RotZ = msg.GetDeatinationRotationZ()
	}
	SomeMapMutex.RUnlock()
}

func (server *Server) translateMessage(msg string) *mc_metadata.P2C_Translate {
	if language, exists := server.langDetector.DetectLanguageOf(msg); exists {
		lng := strings.ToLower(language.String())

		isoCode := constant.GetISOCode(lng)
		if len(isoCode) == 0 {
			fmt.Println("ERROR: Not supported language", lng)
			return nil
		}

		targetIsoCode := constant.GetISOCode("English")
		if isoCode == "en" {
			targetIsoCode = constant.GetISOCode("Korean")
		}

		originID, err := db.FindOriginIDTranslate(msg)
		if err == nil {
			translate, err := server.findTranslate(originID, targetIsoCode)
			if err != nil {
				fmt.Println("ERROR: FindOriginIDTranslate", err)
			}
			if translate != nil {
				return translate
			}
		}

		text, err := translates.GetTranslate(isoCode, targetIsoCode, msg)
		if err != nil || len(text) == 0 {
			fmt.Println("ERROR: GetTranslate", err)
			return nil
		}

		if originID < 0 {
			db.AddTranslate(originID, targetIsoCode, text)
		}

		return &mc_metadata.P2C_Translate{
			Code: targetIsoCode,
			Text: text,
		}
	}
	return nil
}

func (server *Server) convertPlayersToLoginResult(plrs map[string]*mnet.Client) []*mc_metadata.P2C_ReportLoginUser {
	res := make([]*mc_metadata.P2C_ReportLoginUser, 0)

	for _, v := range plrs {
		intr := &mc_metadata.P2C_ReportInteractionAttach{}
		if (*v).GetPlayer().Interaction.IsInteraction {
			intr.AttachEnable = (*v).GetPlayer().Interaction.AttachEnabled
			intr.UuId = (*v).GetPlayer().UId
			intr.ObjectIndex = (*v).GetPlayer().Interaction.ObjectIndex
			intr.AnimMontageName = (*v).GetPlayer().Interaction.AnimMontageName
			intr.DestinationX = (*v).GetPlayer().Interaction.DestinationX
			intr.DestinationY = (*v).GetPlayer().Interaction.DestinationY
			intr.DestinationZ = (*v).GetPlayer().Interaction.DestinationZ
		}

		res = append(res, &mc_metadata.P2C_ReportLoginUser{
			UuId: (*v).GetPlayer().UId,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: (*v).GetPlayer().Character.NickName,
				Hair:     (*v).GetPlayer().Character.Hair,
				Top:      (*v).GetPlayer().Character.Top,
				Bottom:   (*v).GetPlayer().Character.Bottom,
				Clothes:  (*v).GetPlayer().Character.Clothes,
			},
			InteractionData: intr,
			SpawnPosX:       (*v).GetPlayer().Character.PosX,
			SpawnPosY:       (*v).GetPlayer().Character.PosY,
			SpawnPosZ:       (*v).GetPlayer().Character.PosZ,
			SpawnRotX:       (*v).GetPlayer().Character.RotX,
			SpawnRotY:       (*v).GetPlayer().Character.RotY,
			SpawnRotZ:       (*v).GetPlayer().Character.RotZ,
		})
	}
	return res
}

func (server *Server) convertPlayersToRegionReport(plrs map[string]*mnet.Client) []*mc_metadata.P2C_ReportRegionChange {
	var res []*mc_metadata.P2C_ReportRegionChange

	for _, v := range plrs {
		intr := &mc_metadata.P2C_ReportInteractionAttach{}
		if (*v).GetPlayer().Interaction.IsInteraction {
			intr.UuId = (*v).GetPlayer().UId
			intr.AttachEnable = (*v).GetPlayer().Interaction.AttachEnabled
			intr.ObjectIndex = (*v).GetPlayer().Interaction.ObjectIndex
			intr.AnimMontageName = (*v).GetPlayer().Interaction.AnimMontageName
			intr.DestinationX = (*v).GetPlayer().Interaction.DestinationX
			intr.DestinationY = (*v).GetPlayer().Interaction.DestinationY
			intr.DestinationZ = (*v).GetPlayer().Interaction.DestinationZ
		}
		res = append(res, &mc_metadata.P2C_ReportRegionChange{
			UuId:     (*v).GetPlayer().UId,
			RegionId: int32((*v).GetPlayer().RegionID),
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId:     (*v).GetPlayer().UId,
				Role:     (*v).GetPlayer().Character.Role,
				Nickname: (*v).GetPlayer().Character.NickName,
				Hair:     (*v).GetPlayer().Character.Hair,
				Top:      (*v).GetPlayer().Character.Top,
				Bottom:   (*v).GetPlayer().Character.Bottom,
				Clothes:  (*v).GetPlayer().Character.Clothes,
			},
			InteractionData: intr,
			SpawnPosX:       (*v).GetPlayer().Character.PosX,
			SpawnPosY:       (*v).GetPlayer().Character.PosY,
			SpawnPosZ:       (*v).GetPlayer().Character.PosZ,
			SpawnRotX:       (*v).GetPlayer().Character.RotX,
			SpawnRotY:       (*v).GetPlayer().Character.RotY,
			SpawnRotZ:       (*v).GetPlayer().Character.RotZ,
		})
	}
	return res
}

func (server *Server) convertPlayersToGridChanged(plrs map[string]*mnet.Client) []*mc_metadata.GridPlayers {
	res := []*mc_metadata.GridPlayers{}

	for _, v := range plrs {
		intr := &mc_metadata.P2C_ReportInteractionAttach{}
		if (*v).GetPlayer().Interaction.IsInteraction {
			intr.UuId = (*v).GetPlayer().UId
			intr.AttachEnable = (*v).GetPlayer().Interaction.AttachEnabled
			intr.ObjectIndex = (*v).GetPlayer().Interaction.ObjectIndex
			intr.AnimMontageName = (*v).GetPlayer().Interaction.AnimMontageName
			intr.DestinationX = (*v).GetPlayer().Interaction.DestinationX
			intr.DestinationY = (*v).GetPlayer().Interaction.DestinationY
			intr.DestinationZ = (*v).GetPlayer().Interaction.DestinationZ
		}

		res = append(res, &mc_metadata.GridPlayers{

			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId:     (*v).GetPlayer().UId,
				Role:     (*v).GetPlayer().Character.Role,
				Nickname: (*v).GetPlayer().Character.NickName,
				Hair:     (*v).GetPlayer().Character.Hair,
				Top:      (*v).GetPlayer().Character.Top,
				Bottom:   (*v).GetPlayer().Character.Bottom,
				Clothes:  (*v).GetPlayer().Character.Clothes,
			},
			InteractionData: intr,
			SpawnPosX:       (*v).GetPlayer().Character.PosX,
			SpawnPosY:       (*v).GetPlayer().Character.PosY,
			SpawnPosZ:       (*v).GetPlayer().Character.PosZ,
			SpawnRotX:       (*v).GetPlayer().Character.RotX,
			SpawnRotY:       (*v).GetPlayer().Character.RotY,
			SpawnRotZ:       (*v).GetPlayer().Character.RotZ,
		})
	}
	return res
}

func (server *Server) findTranslate(originID int64, lng string) (*mc_metadata.P2C_Translate, error) {
	translate, err := db.GetTranslate(originID, lng)
	if err != nil {
		fmt.Println("ERROR: FindOriginIDTranslate", err)
		return nil, err
	}
	return translate, nil
}

func (server Server) warpPlayer(plr *player, dstField *field, dstPortal portal) error {
	srcField, ok := server.fields[plr.mapID]

	if !ok {
		return fmt.Errorf("Error in map id %d", plr.mapID)
	}

	srcInst, err := srcField.getInstance(plr.inst.id)

	if err != nil {
		return err
	}

	dstInst, err := dstField.getInstance(plr.inst.id)

	if err != nil {
		if dstInst, err = dstField.getInstance(0); err != nil { // Check player is not in higher level instance than available
			return err
		}
	}

	srcInst.removePlayer(plr)

	plr.setMapID(dstField.id)
	// plr.mapPos = dstPortal.id
	plr.pos = dstPortal.pos
	// plr.SetFoothold(0)

	packetMapChange := func(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
		p.WriteInt32(channelID)
		p.WriteByte(0) // character portal counter
		p.WriteByte(0) // Is connecting
		p.WriteInt32(mapID)
		p.WriteByte(mapPos)
		p.WriteInt16(hp)
		p.WriteByte(0) // flag for more reading

		return p
	}

	plr.send(packetMapChange(dstField.id, int32(server.id), dstPortal.id, plr.hp)) // plr.ChangeMap(dstField.ID, dstPortal.ID(), dstPortal.Pos(), foothold)
	dstInst.addPlayer(plr)

	return nil
}

func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelBad:
		server.handleNewChannelBad(conn, reader)
	case opcode.ChannelOk:
		server.handleNewChannelOK(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	case opcode.ChannePlayerConnect:
		server.handlePlayerConnectedNotifications(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnectNotifications(conn, reader)
	case opcode.ChannelPlayerChatEvent:
		server.handleChatEvent(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		server.handleBuddyEvent(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
	case opcode.ChangeRate:
		server.handleChangeRate(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handlePlayerConnectedNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	channelID := reader.ReadByte()
	changeChannel := reader.ReadBool()
	mapID := reader.ReadInt32()

	plr, _ := server.players.getFromID(playerID)

	for _, party := range server.parties {
		party.setPlayerChannel(plr, playerID, false, false, int32(channelID), mapID)
	}

	for i, v := range server.players {
		if v.id == playerID {
			continue
		} else if v.hasBuddy(playerID) {
			if changeChannel {
				server.players[i].send(packetBuddyChangeChannel(playerID, int32(channelID)))
				server.players[i].addOnlineBuddy(playerID, name, int32(channelID))
			} else {
				// send online message card, then update buddy list
				server.players[i].send(packetBuddyOnlineStatus(playerID, int32(channelID)))
				server.players[i].addOnlineBuddy(playerID, name, int32(channelID))
			}
		}
	}
}

func (server *Server) handlePlayerDisconnectNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())

	for _, party := range server.parties {
		party.setPlayerChannel(new(player), playerID, false, true, 0, -1)
	}

	for i, v := range server.players {
		if v.id == playerID {
			continue
		} else if v.hasBuddy(playerID) {
			server.players[i].addOfflineBuddy(playerID, name)
		}
	}
}

func (server *Server) handleBuddyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.send(packetBuddyReceiveRequest(fromID, fromName, int32(channelID)))
	case 2:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.addOfflineBuddy(fromID, fromName)
		plr.send(packetBuddyOnlineStatus(fromID, int32(channelID)))
		plr.addOnlineBuddy(fromID, fromName, int32(channelID))
	case 3:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.removeBuddy(fromID)
	default:
		log.Println("Unknown buddy event type:", op)
	}
}

func (server *Server) handleNewChannelBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by world server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithWorld()
}

func (server *Server) handleNewChannelOK(conn mnet.Server, reader mpacket.Reader) {
	server.worldName = reader.ReadString(reader.ReadInt16())
	server.id = reader.ReadByte()
	server.rates.exp = reader.ReadFloat32()
	server.rates.drop = reader.ReadFloat32()
	server.rates.mesos = reader.ReadFloat32()

	log.Printf("Registered as channel %d on world %s with rates: Exp - x%.2f, Drop - x%.2f, Mesos - x%.2f",
		server.id, server.worldName, server.rates.exp, server.rates.drop, server.rates.mesos)

	for _, p := range server.players {
		p.send(packetMessageNotice("Re-connected to world server as channel " + strconv.Itoa(int(server.id+1))))
		// TODO send largest party id for world server to compare
	}

	accountIDs, err := db.Maria.Query("SELECT accountID from characters where channelID = ? and migrationID = -1", server.id)

	if err != nil {
		log.Println(err)
		return
	}

	for accountIDs.Next() {
		var accountID int
		err := accountIDs.Scan(&accountID)

		if err != nil {
			continue
		}

		_, err = db.Maria.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", 0, accountID)

		if err != nil {
			log.Println(err)
			return
		}
	}

	accountIDs.Close()

	_, err = db.Maria.Exec("UPDATE characters SET channelID=? WHERE channelID=?", -1, server.id)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Loged out any accounts still connected to this channel")
}

func (server *Server) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0:
		log.Println("Channel server should not receive party event message type: 0")
	case 1: // new party created
		channelID := reader.ReadByte()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		// TODO: Mystic door information needs to be sent here if the leader has an active door

		newParty := newParty(partyID, plr, channelID, playerID, mapID, job, level, name, int32(server.id))

		server.parties[partyID] = &newParty

		if plr != nil {
			plr.party = &newParty
			plr.send(packetPartyCreate(1, -1, -1, newPos(0, 0, 0)))
		}
	case 2: // leave party
		destroy := reader.ReadBool()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.removePlayer(plr, playerID, false)

			if destroy {
				delete(server.parties, partyID)
			}
		}
	case 3: // accept
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		channelID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.addPlayer(plr, channelID, playerID, name, mapID, job, level)
		}
	case 4: // expel
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.removePlayer(plr, playerID, true)
		}
	case 5:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		reader.ReadString(reader.ReadInt16()) // name
		if party, ok := server.parties[partyID]; ok {
			party.updateJobLevel(playerID, job, level)
		}
	default:
		log.Println("Unknown party event type:", op)
	}
}

func (server *Server) handleChangeRate(conn mnet.Server, reader mpacket.Reader) {
	mode := reader.ReadByte()
	rate := reader.ReadFloat32()

	modeMap := map[byte]string{
		1: "exp",
		2: "drop",
		3: "mesos",
	}
	switch mode {
	case 1:
		server.rates.exp = rate
	case 2:
		server.rates.drop = rate
	case 3:
		server.rates.mesos = rate
	default:
		log.Println("Unknown rate mode")
		return
	}

	log.Printf("%s rate has changed to x%.2f", modeMap[mode], rate)
	for _, p := range server.players {
		p.conn.Send(packetMessageNotice(fmt.Sprintf("%s rate has changed to x%.2f", modeMap[mode], rate)))
	}

}

func (server Server) handleChatEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0: // whisper
		recepientName := reader.ReadString(reader.ReadInt16())
		fromName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		plr, err := server.players.getFromName(recepientName)

		if err != nil {
			return
		}

		plr.send(packetMessageWhisper(fromName, msg, channelID))

	case 1: // buddy
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.send(packetMessageBubblessChat(0, fromName, msg))
		}
	case 2: // party
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.send(packetMessageBubblessChat(1, fromName, msg))
		}
	case 3: // guild
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.send(packetMessageBubblessChat(2, fromName, msg))
		}
	default:
		log.Println("Unknown chat event type:", op)
	}
}
