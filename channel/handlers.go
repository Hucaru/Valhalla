package channel

import (
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"log"
	rand2 "math/rand"
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

func (server *Server) clientBot(conn *mnet.Client, reader mpacket.Reader) {
	msg := mc_metadata.C2P_Request_BOT{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetNickname()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	switch msg.GetActionType() {
	case 0:
		packet := mc_metadata.C2P_RequestPlayerInfo{
			Nickname: msg.Nickname,
		}

		sendData, _ := proto.MakeResponse(&packet, constant.C2P_RequestPlayerInfo)
		data := mpacket.Packet{}
		data.Append(sendData)
		reader := mpacket.NewReader(&data, time.Now().Unix())
		server.playerInfo(conn, reader)
	case 1:
		packet := mc_metadata.C2P_RequestLoginUser{
			UuId:      conn.GetPlayer().UId,
			IsBot:     1,
			SpawnPosX: msg.SpawnPosX,
			SpawnPosY: msg.SpawnPosY,
			SpawnPosZ: msg.SpawnPosZ,
		}

		sendData, _ := proto.MakeResponse(&packet, constant.C2P_RequestLoginUser)
		data := mpacket.Packet{}
		data.Append(sendData)
		reader := mpacket.NewReader(&data, time.Now().Unix())
		server.playerConnect(conn, reader)
	case 2:
		MovementData := mc_metadata.Movement{}
		MovementData.UuId = conn.GetPlayer().UId
		MovementData.DestinationX = msg.SpawnPosX
		MovementData.DestinationY = msg.SpawnPosY
		MovementData.DestinationZ = msg.SpawnPosZ
		MovementData.InterpTime = 320
		packet := mc_metadata.C2P_RequestMoveStart{
			MovementData: &MovementData,
		}

		sendData, _ := proto.MakeResponse(&packet, constant.C2P_RequestMoveStart)
		data := mpacket.Packet{}
		data.Append(sendData)
		reader := mpacket.NewReader(&data, time.Now().Unix())
		server.playerMovementStart(conn, reader)
	case 3:
		MovementData := mc_metadata.Movement{}
		MovementData.UuId = conn.GetPlayer().UId
		MovementData.DestinationX = msg.SpawnPosX
		MovementData.DestinationY = msg.SpawnPosY
		MovementData.DestinationZ = msg.SpawnPosZ
		MovementData.InterpTime = 320
		packet := mc_metadata.C2P_RequestMove{
			MovementData: &MovementData,
		}

		sendData, _ := proto.MakeResponse(&packet, constant.C2P_RequestMove)
		data := mpacket.Packet{}
		data.Append(sendData)
		reader := mpacket.NewReader(&data, time.Now().Unix())
		server.playerMovementStart(conn, reader)
	case 4:
		MovementData := mc_metadata.Movement{}
		MovementData.UuId = conn.GetPlayer().UId
		MovementData.DestinationX = msg.SpawnPosX
		MovementData.DestinationY = msg.SpawnPosY
		MovementData.DestinationZ = msg.SpawnPosZ
		MovementData.InterpTime = 320
		packet := mc_metadata.C2P_RequestMoveEnd{
			MovementData: &MovementData,
		}

		sendData, _ := proto.MakeResponse(&packet, constant.C2P_RequestMoveEnd)
		data := mpacket.Packet{}
		data.Append(sendData)
		reader := mpacket.NewReader(&data, time.Now().Unix())
		server.playerMovementStart(conn, reader)
	}
}

// HandleClientPacket data
func (server *Server) HandleClientPacket(
	conn *mnet.Client, reader mpacket.Reader, msgProtocolType uint32) {

	f, ok := server.PlayerActionHandler[msgProtocolType]
	if ok {
		f(conn, reader)
	}
}

func (server *Server) playerConnect(conn *mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestLoginUser{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || msg.GetUuId() == 0 {
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
			err1 := db.UpdateLoginState(player.UId, true)
			if err1 != nil {
				log.Println("Unable to complete login for ", player.UId)
				m, err2 := proto.ErrorLoginResponse(err.Error(), player.UId)
				if err2 != nil {
					log.Println("ErrorLoginResponse", err2)
				}
				conn.BaseConn.Send(m)
				return
			}
		}
	}

	res := &mc_metadata.P2C_ResultPlayerInfo{
		ErrorCode: constant.NoError,
	}

	if server.isPlayerOnline(msg.GetUuId()) {
		res.UId = msg.GetUuId()
		res.ErrorCode = constant.ErrorCodeAlreadyOnline

		data, err := proto.MakeResponse(res, constant.P2C_ResultPlayerInfo)
		if err != nil {
			log.Println("ERROR P2C_ResultPlayerInfo Already Online", msg.GetUuId())
			return
		}
		conn.Send(data)
		return
	}

	ch := player.GetCharacter_P()

	if msg.GetSpawnPosX() != 0 {
		ch.PosX = msg.GetSpawnPosX()
	}

	if msg.GetSpawnPosY() != 0 {
		ch.PosY = msg.GetSpawnPosY()
	}

	if msg.GetSpawnPosZ() != 0 {
		ch.PosZ = msg.GetSpawnPosZ()
	}

	player.IsBot = msg.IsBot

	// TMP part, will be moved later
	if player.RegionID == constant.MetaClassRoom {
		player.RegionID = constant.MetaSchool
		ch.PosX = -8597
		ch.PosY = -23392
		ch.PosZ = 2180

		db.UpdateRegionID(player.CharacterID, int32(player.RegionID))
	}

	res.UId = player.UId

	//plr := loadPlayer(conn, *msg)
	//plr.rates = &server.rates
	//plr.conn.SetPlayer(*player)
	//
	//server.addPlayer(&plr)
	conn.SetPlayer(*player)
	conn.TempIsBot = msg.IsBot == 1
	ch = conn.GetPlayer_P().GetCharacter_P()

	if msg.IsBot == 1 {
		ch.Top = constant.RandomTop[rand2.Intn(4)]
		ch.Bottom = constant.RandomBottom[rand2.Intn(4)]
		ch.Clothes = constant.RandomClothes[rand2.Intn(4)]
		ch.Hair = constant.RandomHair[rand2.Intn(5)]
		//go server.addToEmulateMove(&plr)
		//return
	}

	GridX, GridY := common.FindGrid(ch.PosX, ch.PosY)

	server.gridMgr.Add(conn.GetPlayer().RegionID, GridX, GridY, conn)
	server.clients.Set(conn.GetPlayer().UId, conn)

	reportLoginUserPacket := mc_metadata.P2C_ReportLoginUser{
		UuId: player.UId,
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: ch.NickName,
			Hair:     ch.Hair,
			Top:      ch.Top,
			Bottom:   ch.Bottom,
			Clothes:  ch.Clothes,
		},
		SpawnPosX: ch.PosX,
		SpawnPosY: ch.PosY,
		SpawnPosZ: ch.PosZ,
		SpawnRotX: ch.RotX,
		SpawnRotY: ch.RotY,
		SpawnRotZ: ch.RotZ,
	}

	server.sendMsgToRegion(conn, &reportLoginUserPacket, constant.P2C_ReportLoginUser)
	_, _, newList := server.gridMgr.OnMove(player.RegionID, ch.PosX, ch.PosY, player.UId)
	//

	resultLoginUserPacket := proto.AccountResult(player)
	for _, v := range newList {
		_v := v.GetPlayer()
		_ch := _v.GetCharacter()
		PlayerInfo := mc_metadata.P2C_PlayerInfo{
			Nickname: _ch.NickName,
			Hair:     _ch.Hair,
			Top:      _ch.Top,
			Bottom:   _ch.Bottom,
			Clothes:  _ch.Clothes,
		}
		_r := mc_metadata.P2C_ReportLoginUser{
			UuId:       _v.UId,
			PlayerInfo: &PlayerInfo,
			SpawnPosX:  _ch.PosX,
			SpawnPosY:  _ch.PosY,
			SpawnPosZ:  _ch.PosZ,
			SpawnRotX:  _ch.RotX,
			SpawnRotY:  _ch.RotY,
			SpawnRotZ:  _ch.RotZ,
		}
		resultLoginUserPacket.LoggedUsers = append(resultLoginUserPacket.LoggedUsers, &_r)
	}

	server.sendMsgToMe(conn, &resultLoginUserPacket, constant.P2C_ResultLoginUser)
}

func (server *Server) getRoomPlayers(uID int64, mX, mY float32) []*model.Player {
	plrs := make([]*model.Player, 0)
	return plrs
}

func (server *Server) setPlayer(plr *model.Player) {
}

func (server *Server) isPlayerOnline(uID int64) bool {
	return server.clients.Has(uID)
}

func (server *Server) sendMsgToMe(conn *mnet.Client, msg proto2.Message, msgType int) {
	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	if err != nil {
		return
	}
	conn.Send(res)
}

func (server *Server) sendMsgToPlayer(msg proto2.Message, uID int64, msgType int) {

	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	c, ok := server.clients.Get(uID)
	if ok {
		c.Send(res)
	}
}

func (server *Server) sendMsgToAll(msg proto2.Message, uID int64, msgType int) {
	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	server.clients.IterCb(func(k int64, v *mnet.Client) {
		if k == uID {
			return
		}

		v.Send(res)
	})
}

func (server *Server) sendMsgToRegion(conn *mnet.Client, msg proto2.Message, msgType int) {
	x, y := common.FindGrid(conn.GetPlayer_P().GetCharacter().PosX, conn.GetPlayer_P().GetCharacter().PosY)
	plrs := server.getPlayersOnGrids(conn.GetPlayer().RegionID, x, y, conn.GetPlayer().UId)

	//log.Println("getPlayersOnGrids", len(plrs))

	res, err := proto.MakeResponse(msg, uint32(msgType))
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}

	uId := conn.GetPlayer().UId

	for _, p := range plrs {
		if uId == p.GetPlayer().UId {
			continue
		}
		p.Send(res)
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

	msg := mc_metadata.C2P_RequestMoveStart{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || msg.GetMovementData().GetUuId() == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMoveStart)
}

func (server *Server) playerMovementEnd(conn *mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMoveEnd{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || msg.GetMovementData().GetUuId() == 0 {
		log.Println("Failed to parse data:", err)
		return
	}
	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMoveEnd)
}

func (server *Server) getPlayersOnGrids(regionId int64, x, y int, uID int64) map[int64]*mnet.Client {
	oldList := map[int64]*mnet.Client{}

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			oldGridX := x + i
			oldGridY := y + j

			maps.Copy(oldList, server.gridMgr.FillPlayers(regionId, oldGridX, oldGridY))
		}
	}

	delete(oldList, uID)
	return oldList
}

func (server *Server) existsPlayerFromGrid(uID int64, x, y int) bool {
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
	att := &mc_metadata.P2C_ResultInteractionAttach{
		ErrorCode: -1,
	}

	p := conn.GetPlayer_P()
	interaction := p.GetInteraction_P()

	if p.GetInteraction().IsInteraction == false {
		p.SetInteraction(model.NewInteraction())
	}
	interaction.ObjectIndex = msg.GetObjectIndex()
	interaction.AttachEnabled = msg.GetAttachEnable()
	interaction.AnimMontageName = msg.GetAnimMontageName()
	interaction.DestinationX = msg.GetDestinationX()
	interaction.DestinationY = msg.GetDestinationY()
	interaction.DestinationZ = msg.GetDestinationZ()

	server.sendMsgToMe(conn, att, constant.P2C_ResultInteractionAttach)
	return nil
}

func (server *Server) InsertInteractionAndSend(conn *mnet.Client, msg *mc_metadata.C2P_RequestInteractionAttach) error {
	att := &mc_metadata.P2C_ResultInteractionAttach{
		ErrorCode: -1,
	}

	p := conn.GetPlayer_P()

	x, y := common.FindGrid(p.GetCharacter().PosX, p.GetCharacter().PosX)
	users := server.getGridPlayers(x, y)

	for i := 0; i < len(users); i++ {
		u := users[i].conn.GetPlayer_P()
		if u.GetInteraction().IsInteraction == false &&
			conn.GetPlayer().UId != users[i].conn.GetPlayer().UId &&
			msg.ObjectIndex == u.GetInteraction().ObjectIndex {
			att.ErrorCode = constant.ErrorCodeChairNotEmpty
			break
		}
	}

	if att.ErrorCode == -1 {
		if p.GetInteraction().IsInteraction == false {
			p.SetInteraction(model.NewInteraction())
		}

		interaction := p.GetInteraction_P()

		interaction.ObjectIndex = msg.GetObjectIndex()
		interaction.AttachEnabled = msg.GetAttachEnable()
		interaction.AnimMontageName = msg.GetAnimMontageName()
		interaction.DestinationX = msg.GetDestinationX()
		interaction.DestinationY = msg.GetDestinationY()
		interaction.DestinationZ = msg.GetDestinationZ()
	}

	server.sendMsgToMe(conn, att, constant.P2C_ResultInteractionAttach)
	if att.ErrorCode == -1 {
		return nil
	} else {
		return errors.New("chair not empty")
	}

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

	x, y := common.FindGrid(conn.GetPlayer_P().GetCharacter().PosX, conn.GetPlayer_P().GetCharacter().PosX)
	users := server.getGridPlayers(x, y)

	is := 0
	for i := 0; i < len(users); i++ {
		p := users[i].conn.GetPlayer()
		if p.GetInteraction().IsInteraction &&
			users[i].conn.GetPlayer_P().GetCharacter().Role > 1 {
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

	msg := mc_metadata.C2P_RequestMetaSchoolEnter{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	vPlayer := conn.GetPlayer()
	pPlayer := conn.GetPlayer_P()
	pCh := pPlayer.GetCharacter_P()
	vCh := pPlayer.GetCharacter()

	pCh.Role = msg.TeacherEnable

	interaction := model.NewInteraction()
	interaction.AttachEnabled = 1
	interaction.ObjectIndex = -1
	conn.GetPlayer_P().SetInteraction(interaction)

	reportEnter := mc_metadata.P2C_ReportMetaSchoolEnter{
		UuId:          msg.GetUuId(),
		TeacherEnable: msg.GetTeacherEnable(),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			UuId:     vPlayer.UId,
			Nickname: vCh.NickName,
			Role:     vCh.Role,
			Hair:     vCh.Hair,
			Top:      vCh.Top,
			Bottom:   vCh.Bottom,
			Clothes:  vCh.Clothes,
		},
	}

	log.Println("P2C_ResultMetaSchoolEnter sendMsgToRegion")
	go server.sendMsgToRegion(conn, &reportEnter, constant.P2C_ReportMetaSchoolEnter)

	res := mc_metadata.P2C_ResultMetaSchoolEnter{
		UuId:          msg.GetUuId(),
		TeacherEnable: msg.GetTeacherEnable(),
		DataSchool:    proto.ConvertPlayersToRoomReport(server.getRoomPlayers(msg.GetUuId(), vCh.PosX, vCh.PosY)),
	}

	server.sendMsgToMe(conn, &res, constant.P2C_ResultMetaSchoolEnter)
}

func (server *Server) playerLeaveFromRoom(conn *mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMetaSchoolLeave{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	vPlayer := conn.GetPlayer()
	pPlayer := conn.GetPlayer_P()
	vCh := vPlayer.GetCharacter()
	pPlayer.GetInteraction_P().IsInteraction = false

	res := &mc_metadata.P2C_ReportMetaSchoolLeave{
		UuId: msg.GetUuId(),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			UuId:     vPlayer.UId,
			Nickname: vCh.NickName,
			Role:     vCh.Role,
			Hair:     vCh.Hair,
			Top:      vCh.Top,
			Bottom:   vCh.Bottom,
			Clothes:  vCh.Clothes,
		},
	}
	go server.sendMsgToRegion(conn, res, constant.P2C_ReportMetaSchoolLeave)
}

func (server *Server) playerMovement(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestMove{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || msg.GetMovementData().GetUuId() == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	server.moveProcess(conn, msg.GetMovementData().DestinationX, msg.GetMovementData().DestinationY, msg.GetMovementData().GetUuId(), msg.GetMovementData(), constant.P2C_ReportMove)
}

func (server *Server) moveProcess(conn *mnet.Client, x, y float32, uId int64, movement *mc_metadata.Movement, moveType int) {
	addList, removeList, aroundList := server.gridMgr.OnMove(conn.GetPlayer().RegionID, x, y, uId)

	for _, v := range addList {
		c := v.GetPlayer_P().GetCharacter()

		__PlayerInfo := mc_metadata.P2C_PlayerInfo{
			Nickname: c.NickName,
			UuId:     uId,
			Top:      c.Top,
			Bottom:   c.Bottom,
			Clothes:  c.Clothes,
			Hair:     c.Hair,
		}

		res := mc_metadata.P2C_ReportGridNew{
			PlayerInfo: &__PlayerInfo,
			SpawnPosX:  c.PosX,
			SpawnPosY:  c.PosY,
			SpawnPosZ:  c.PosZ,
			SpawnRotX:  c.RotX,
			SpawnRotY:  c.RotY,
			SpawnRotZ:  c.RotZ,
		}

		p := conn.GetPlayer_P()
		ch := p.GetCharacter()

		_PlayerInfo := mc_metadata.P2C_PlayerInfo{
			UuId:     p.UId,
			Nickname: ch.NickName,
			Top:      ch.Top,
			Bottom:   ch.Bottom,
			Clothes:  ch.Clothes,
			Hair:     ch.Hair,
		}

		res2 := mc_metadata.P2C_ReportGridNew{
			PlayerInfo: &_PlayerInfo,
			SpawnPosX:  ch.PosX,
			SpawnPosY:  ch.PosY,
			SpawnPosZ:  ch.PosZ,
			SpawnRotX:  ch.RotX,
			SpawnRotY:  ch.RotY,
			SpawnRotZ:  ch.RotZ,
		}

		server.sendMsgToMe(conn, &res, constant.P2C_ReportGridNew)
		server.sendMsgToMe(v, &res2, constant.P2C_ReportGridNew)
	}

	for k, v := range removeList {
		p1 := mc_metadata.P2C_PlayerInfo{
			UuId: k,
		}

		res := mc_metadata.P2C_ReportGridOld{
			PlayerInfo: &p1,
		}

		p2 := mc_metadata.P2C_PlayerInfo{
			UuId: conn.GetPlayer().UId,
		}

		res2 := mc_metadata.P2C_ReportGridOld{
			PlayerInfo: &p2,
		}

		//fmt.Println(fmt.Sprintf("conn : %s v : %s res : %s res2 : %s", conn.GetPlayer().UId, v.GetPlayer().UId, res.PlayerInfo.UuId, res2.PlayerInfo.UuId))

		server.sendMsgToMe(conn, &res, constant.P2C_ReportGridOld)
		server.sendMsgToMe(v, &res2, constant.P2C_ReportGridOld)
	}

	switch moveType {
	case constant.P2C_ReportMoveStart:
		res := mc_metadata.P2C_ReportMoveStart{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, &res, constant.P2C_ReportMoveStart)
		}
	case constant.P2C_ReportMove:
		res := mc_metadata.P2C_ReportMove{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, &res, constant.P2C_ReportMove)
		}
	case constant.P2C_ReportMoveEnd:
		res := mc_metadata.P2C_ReportMoveEnd{
			MovementData: movement,
		}

		for _, v := range aroundList {
			server.sendMsgToMe(v, &res, constant.P2C_ReportMoveEnd)
		}
	}

	ch := conn.GetPlayer_P().GetCharacter_P()
	ch.PosX = movement.GetDestinationX()
	ch.PosY = movement.GetDestinationY()
	ch.PosZ = movement.GetDestinationZ()
	ch.RotX = movement.GetDeatinationRotationX()
	ch.RotY = movement.GetDeatinationRotationY()
	ch.RotZ = movement.GetDeatinationRotationZ()
}

func (server *Server) playerInfo(conn *mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestPlayerInfo{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetNickname()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := &mc_metadata.P2C_ResultPlayerInfo{

		ErrorCode: constant.NoError,
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

	res.UId = plr.UId
	data, err := proto.MakeResponse(res, constant.P2C_ResultPlayerInfo)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", msg.GetNickname())
		return
	}

	conn.BaseConn.Send(data)

	//server.sendMsgToMe(data, conn)
	data = nil
}

func (server *Server) playerLogout(conn *mnet.Client, reader mpacket.Reader) {

	msg := &mc_metadata.C2P_RequestLogoutUser{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || msg.GetUuId() == 0 {
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
	if err != nil || msg.GetUuId() == 0 {
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
	if err != nil {
		return
	}

	db.AddPublicMessage(conn.GetPlayer().CharacterID, constant.World, msg.GetChat())

	toMe := &mc_metadata.P2C_ResultAllChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	server.sendMsgToMe(conn, toMe, constant.P2C_ResultAllChat)
}

func (server *Server) chatSendRegion(conn *mnet.Client, reader mpacket.Reader) {
	msg := mc_metadata.C2P_RequestRegionChat{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || msg.GetUuId() == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	translate := mc_metadata.P2C_Translate{}
	res := mc_metadata.P2C_ReportRegionChat{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		Translate: &translate,
	}

	papagoTranslate := server.translateMessage(msg.GetChat())
	if papagoTranslate != nil {
		res.Translate = papagoTranslate
	}

	server.sendMsgToRegion(conn, &res, constant.P2C_ReportRegionChat)
	db.AddPublicMessage(
		conn.GetPlayer().CharacterID,
		conn.GetPlayer().RegionID,
		msg.GetChat())

	toMe := mc_metadata.P2C_ResultRegionChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	server.sendMsgToMe(conn, &toMe, constant.P2C_ResultRegionChat)
}

func (server *Server) chatSendWhisper(conn *mnet.Client, reader mpacket.Reader) {
	msg := mc_metadata.C2P_RequestWhisper{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || msg.GetUuId() == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := mc_metadata.P2C_ReportWhisper{
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

	toMe := mc_metadata.P2C_ResultWhisper{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetTargetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		ErrorCode: constant.NoError,
	}

	targetPlayer := server.findCharacterID(msg.GetUuId())
	db.AddWhisperMessage(
		conn.GetPlayer().CharacterID,
		targetPlayer.CharacterID,
		msg.GetChat())

	if targetPlayer != nil {
		server.sendMsgToPlayer(&res, targetPlayer.UId, constant.P2C_ReportWhisper)
	} else {
		toMe.ErrorCode = constant.ErrorUserOffline
	}

	server.sendMsgToMe(conn, &toMe, constant.P2C_ResultWhisper)
}

func (server *Server) findCharacterID(uId int64) *model.Player {
	v, _ := server.clients.Get(uId)
	return v.GetPlayer_P()
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

		if originID > 0 {
			db.AddTranslate(originID, targetIsoCode, text)
		} else {
			id, err1 := db.AddTranslate(-1, isoCode, msg)
			if err1 == nil {
				db.AddTranslate(id, targetIsoCode, text)
			}
		}

		return &mc_metadata.P2C_Translate{
			Code: targetIsoCode,
			Text: text,
		}
	}
	return nil
}

func (server *Server) convertPlayersToLoginResult(plrs map[string]*mnet.Client) []mc_metadata.P2C_ReportLoginUser {
	res := make([]mc_metadata.P2C_ReportLoginUser, 0)

	for _, v := range plrs {
		intr := mc_metadata.P2C_ReportInteractionAttach{}
		p := v.GetPlayer()
		interaction := v.GetPlayer_P().GetInteraction()
		if interaction.IsInteraction {
			intr.AttachEnable = interaction.AttachEnabled
			intr.UuId = p.UId
			intr.ObjectIndex = interaction.ObjectIndex
			intr.AnimMontageName = interaction.AnimMontageName
			intr.DestinationX = interaction.DestinationX
			intr.DestinationY = interaction.DestinationY
			intr.DestinationZ = interaction.DestinationZ
		}

		ch := p.GetCharacter()

		res = append(res, mc_metadata.P2C_ReportLoginUser{
			UuId: p.UId,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: ch.NickName,
				Hair:     ch.Hair,
				Top:      ch.Top,
				Bottom:   ch.Bottom,
				Clothes:  ch.Clothes,
			},
			InteractionData: &intr,
			SpawnPosX:       ch.PosX,
			SpawnPosY:       ch.PosY,
			SpawnPosZ:       ch.PosZ,
			SpawnRotX:       ch.RotX,
			SpawnRotY:       ch.RotY,
			SpawnRotZ:       ch.RotZ,
		})
	}
	return res
}

func (server *Server) convertPlayersToRegionReport(plrs map[string]*mnet.Client) []*mc_metadata.P2C_ReportRegionChange {
	var res []*mc_metadata.P2C_ReportRegionChange

	for _, v := range plrs {
		intr := &mc_metadata.P2C_ReportInteractionAttach{}
		p := v.GetPlayer()
		ch := p.GetCharacter()
		interaction := v.GetPlayer_P().GetInteraction()

		if interaction.IsInteraction {
			intr.UuId = p.UId
			intr.AttachEnable = interaction.AttachEnabled
			intr.ObjectIndex = interaction.ObjectIndex
			intr.AnimMontageName = interaction.AnimMontageName
			intr.DestinationX = interaction.DestinationX
			intr.DestinationY = interaction.DestinationY
			intr.DestinationZ = interaction.DestinationZ
		}
		res = append(res, &mc_metadata.P2C_ReportRegionChange{
			UuId:     p.UId,
			RegionId: int32(p.RegionID),
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId:     p.UId,
				Role:     ch.Role,
				Nickname: ch.NickName,
				Hair:     ch.Hair,
				Top:      ch.Top,
				Bottom:   ch.Bottom,
				Clothes:  ch.Clothes,
			},
			InteractionData: intr,
			SpawnPosX:       ch.PosX,
			SpawnPosY:       ch.PosY,
			SpawnPosZ:       ch.PosZ,
			SpawnRotX:       ch.RotX,
			SpawnRotY:       ch.RotY,
			SpawnRotZ:       ch.RotZ,
		})
	}
	return res
}

func (server *Server) convertPlayersToGridChanged(plrs map[string]*mnet.Client) []*mc_metadata.GridPlayers {
	res := []*mc_metadata.GridPlayers{}

	//for _, v := range plrs {
	//	intr := &mc_metadata.P2C_ReportInteractionAttach{}
	//	if v.GetPlayer().GetInteraction().IsInteraction {
	//		intr.UuId = v.GetPlayer().UId
	//		intr.AttachEnable = v.GetPlayer().GetInteraction().AttachEnabled
	//		intr.ObjectIndex = v.GetPlayer().GetInteraction().ObjectIndex
	//		intr.AnimMontageName = v.GetPlayer().GetInteraction().AnimMontageName
	//		intr.DestinationX = v.GetPlayer().GetInteraction().DestinationX
	//		intr.DestinationY = v.GetPlayer().GetInteraction().DestinationY
	//		intr.DestinationZ = v.GetPlayer().GetInteraction().DestinationZ
	//	}
	//
	//	res = append(res, &mc_metadata.GridPlayers{
	//
	//		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
	//			UuId:     v.GetPlayer().UId,
	//			Role:     v.GetPlayer().GetCharacter().Role,
	//			Nickname: v.GetPlayer().GetCharacter().NickName,
	//			Hair:     v.GetPlayer().GetCharacter().Hair,
	//			Top:      v.GetPlayer().GetCharacter().Top,
	//			Bottom:   v.GetPlayer().GetCharacter().Bottom,
	//			Clothes:  v.GetPlayer().GetCharacter().Clothes,
	//		},
	//		InteractionData: intr,
	//		SpawnPosX:       v.GetPlayer().GetCharacter().PosX,
	//		SpawnPosY:       v.GetPlayer().GetCharacter().PosY,
	//		SpawnPosZ:       v.GetPlayer().GetCharacter().PosZ,
	//		SpawnRotX:       v.GetPlayer().GetCharacter().RotX,
	//		SpawnRotY:       v.GetPlayer().GetCharacter().RotY,
	//		SpawnRotZ:       v.GetPlayer().GetCharacter().RotZ,
	//	})
	//}
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
