package channel

import (
	"errors"
	"fmt"
	"github.com/Hucaru/Valhalla/common/db/model"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/common/db"
	"github.com/Hucaru/Valhalla/common/proto"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	proto2 "google.golang.org/protobuf/proto"

	"github.com/Hucaru/Valhalla/internal"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

// HandleClientPacket data
func (server *Server) HandleClientPacket(conn mnet.Client, tcpConn net.Conn, reader mpacket.Reader, msgProtocolType uint32) {

	switch msgProtocolType {
	case constant.C2P_RequestLoginUser:
		log.Println("PLAYERS ONLINE ", len(server.players))
		server.playerConnect(conn, tcpConn, reader)
		break
	case constant.C2P_RequestMoveStart:
		log.Println("C2P_RequestMoveStart")
		server.playerMovementStart(conn, reader)
		break
	case constant.C2P_RequestMove:
		server.playerMovement(conn, reader)
		break
	case constant.C2P_RequestMoveEnd:
		log.Println("C2P_RequestMoveEnd")
		server.playerMovementEnd(conn, reader)
		break
	case constant.C2P_RequestLogoutUser:
		log.Println("DATA_BUFFER_LOGOUT", reader.GetBuffer())
		server.playerLogout(conn, reader)
		break
	case constant.C2P_RequestPlayerInfo:
		log.Println("DATA_PLAYER_INFO", reader.GetBuffer())
		server.playerInfo(conn, reader)
		break
	case constant.C2P_RequestAllChat:
		log.Println("DATA_ALL_CHAT", reader.GetBuffer())
		go server.chatSendAll(conn, reader)
		break
	case constant.C2P_RequestWhisper:
		log.Println("DATA_WHISPER_CHAT", reader.GetBuffer())
		go server.chatSendWhisper(conn, reader)
		break
	case constant.C2P_RequestRegionChat:
		log.Println("DATA_REGION_CHAT", reader.GetBuffer())
		go server.chatSendRegion(conn, reader)
		break
	case constant.C2P_RequestRegionChange:
		log.Println("DATA_REGION_CHANGE", reader.GetBuffer())
		server.playerChangeChannel(conn, reader)
		break
	case constant.C2P_RequestInteractionAttach:
		log.Println("DATA_INTERACTION", reader.GetBuffer())
		go server.playerInteraction(conn, reader)
		break
	case constant.C2P_RequestPlayMontage:
		log.Println("DATA_PLAY_MONTAGE", reader.GetBuffer())
		server.playerPlayAnimation(conn, reader)
		break
	case constant.C2P_RequestMetaSchoolEnter:
		log.Println("DATA_META_SCHOOL_ENTER", reader.GetBuffer())
		server.playerEnterToRoom(conn, reader)
		break
	case constant.C2P_RequestMetaSchoolLeave:
		log.Println("DATA_META_SCHOOL_ENTER", reader.GetBuffer())
		server.playerLeaveFromRoom(conn, reader)
		break
	case constant.C2P_RequestRoleChecking:
		log.Println("DATA_ROLE_CHECKING", reader.GetBuffer())
		go server.playerRegionRoleChecking(conn, reader)
		break
	default:
		fmt.Println("UNKNOWN MSG", reader)
		//msg = nil
		break
	}
}

func (server *Server) playerConnect(conn mnet.Client, tcpConn net.Conn, reader mpacket.Reader) {
	msg := mc_metadata.C2P_RequestLoginUser{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.UuId) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	player, err := db.GetLoggedData(msg.GetUuId())

	if err != nil {
		log.Println("Inserting new user", msg.GetUuId())
		db.InsertNewAccount(player)
	} else {
		err1 := db.UpdateLoginState(msg.GetUuId(), true)
		if err1 != nil {
			log.Println("Unable to complete login for ", msg.GetUuId())
			m, err2 := proto.ErrorLoginResponse(err.Error(), msg.GetUuId())
			if err2 != nil {
				log.Println("ErrorLoginResponse", err2)
			}
			tcpConn.Write(m)
			return
		}
	}

	// TMP part, will be moved later
	if player.RegionID == constant.MetaClassRoom {
		player.RegionID = constant.MetaSchool
		player.Character.PosX = -8597
		player.Character.PosY = -23392
		player.Character.PosZ = 2180

		db.UpdateRegionID(player.CharacterID, int32(player.RegionID))
	}

	plr := loadPlayer(conn, msg)
	plr.rates = &server.rates
	plr.conn.SetPlayer(*player)
	server.players = append(server.players, &plr)

	response := proto.AccountReport(&player.UId, player.Character)
	res, err := proto.MakeResponse(response, constant.P2C_ReportLoginUser)
	if err != nil {
		log.Println("DATA_RESPONSE_ERROR", err)
	}
	log.Println("PLAYER_ID_LOGIN", player.UId)
	server.sendMsgToAll(res, msg.GetUuId())

	account := proto.AccountResult(player)
	loggedPlayers := server.getLoggedPlayers(player.UId, player.RegionID)

	if loggedPlayers != nil {
		users := proto.ConvertPlayersToLoginResult(loggedPlayers)
		account.LoggedUsers = append(account.LoggedUsers, users...)
	}

	data, err := proto.MakeResponse(account, constant.P2C_ResultLoginUser)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", plr.playerID)
		return
	}

	server.sendMsgToMe(data, conn)
	res = nil
	response = nil
	data = nil
	account = nil
}

func (server *Server) getLoggedPlayers(uID string, regionID int64) []*model.Player {
	plrs := make([]*model.Player, 0)
	for i := 0; i < len(server.players); i++ {
		if regionID != -1 && regionID != server.players[i].conn.GetPlayer().RegionID {
			continue
		}
		if uID == server.players[i].conn.GetPlayer().UId {
			continue
		}

		if uID == server.players[i].conn.GetPlayer().UId {
			continue
		}

		if regionID != server.players[i].conn.GetPlayer().RegionID {
			continue
		}

		plrs = append(plrs, server.players[i].conn.GetPlayer())
	}
	return plrs
}

func (server *Server) getRoomPlayers(uID string) []*model.Player {
	plrs := make([]*model.Player, 0)
	for i := 0; i < len(server.players); i++ {
		if uID == server.players[i].conn.GetPlayer().UId {
			continue
		}
		if server.players[i].conn.GetPlayer().Interaction != nil {
			plrs = append(plrs, server.players[i].conn.GetPlayer())
		}

	}
	return plrs
}

func (server *Server) setPlayer(plr *model.Player) {
	for i := 0; i < len(server.players); i++ {
		if plr.UId == server.players[i].conn.GetPlayer().UId {
			server.players[i].conn.SetPlayer(*plr)
			break
		}
	}
}

func (server *Server) isPlayerOnline(nickname string, uID string) bool {

	for i := 0; i < len(server.players); i++ {
		if nickname == server.players[i].conn.GetPlayer().Character.NickName ||
			uID == server.players[i].conn.GetPlayer().UId {
			return true
		}
	}
	return false
}

func (server *Server) sendMsgToMe(res mpacket.Packet, conn mnet.Client) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}
	plr.conn.Send(res)
}

func (server *Server) sendMsgToPlayer(res mpacket.Packet, uID string) {
	for i := 0; i < len(server.players); i++ {
		if uID == server.players[i].conn.GetPlayer().UId {
			log.Println("sendMsgToPlayer PLAYER_ID", server.players[i].playerID)
			server.players[i].conn.Send(res)
			break
		}
	}
}

func (server *Server) sendMsgToAll(res mpacket.Packet, uID string) {
	for i := 0; i < len(server.players); i++ {
		if uID != server.players[i].conn.GetPlayer().UId {
			server.players[i].conn.Send(res)
		}
	}
}

func (server *Server) sendMsgToRegion(res mpacket.Packet, uID string, regionId int64) {
	for i := 0; i < len(server.players); i++ {
		if uID != server.players[i].conn.GetPlayer().UId &&
			regionId == server.players[i].conn.GetPlayer().RegionID {
			server.players[i].conn.Send(res)
		}
	}
}

func (server *Server) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	msg := mc_metadata.C2P_RequestRegionChange{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.UuId) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	db.UpdateRegionID(plr.conn.GetPlayer().CharacterID, msg.GetRegionId())

	responseOld := proto.ChannelChangeForOldReport(plr.conn.GetPlayer().UId, plr.conn.GetPlayer().Character)
	res1, err1 := proto.MakeResponse(responseOld, constant.P2C_ReportRegionLeave)
	if err1 != nil {
		log.Println("DATA_RESPONSE_ERROR", err1)
	}

	log.Println("REGION_CHANGED PREV REGION SEND", plr.conn.GetPlayer().RegionID)
	server.sendMsgToRegion(res1, plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)

	responseNew := proto.ChannelChangeForNewReport(plr.conn.GetPlayer())
	res2, err2 := proto.MakeResponse(responseNew, constant.P2C_ReportRegionChange)
	if err2 != nil {
		log.Println("DATA_RESPONSE_ERROR", err2)
	}

	log.Println("REGION_CHANGED TO ", msg.GetRegionId())
	server.sendMsgToRegion(res2, plr.conn.GetPlayer().UId, int64(msg.RegionId))

	plr.conn.GetPlayer().RegionID = int64(msg.RegionId)
	server.setPlayer(plr.conn.GetPlayer())

	account := proto.RegionResult(plr.conn.GetPlayer())
	loggedAccounts := server.getLoggedPlayers(plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)

	if err != nil {
		log.Println("ERROR GetLoggedUsersData", plr.conn.GetPlayer().UId)
		return
	}

	users := proto.ConvertPlayersToRegionReport(loggedAccounts)
	account.RegionUsers = append(account.RegionUsers, users...)

	data, err := proto.MakeResponse(account, constant.P2C_ResultRegionChange)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", err)
		return
	}

	server.sendMsgToMe(data, conn)
}

func (server *Server) playerMovementStart(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMoveStart{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ReportMoveStart{
		MovementData: proto.MakeMovementData(msg.GetMovementData()),
	}
	server.makeReportToRegion(conn, &res, constant.P2C_ReportMoveStart)
	go server.updateUserLocation(msg.GetMovementData())
}

func (server *Server) playerMovementEnd(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMoveEnd{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ReportMoveEnd{
		MovementData: proto.MakeMovementData(msg.GetMovementData()),
	}

	server.makeReportToRegion(conn, &res, constant.P2C_ReportMoveEnd)
	go server.updateUserLocation(msg.GetMovementData())
}

func (server *Server) playerInteraction(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestInteractionAttach{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	errR := errors.New("error")
	errR = nil

	if msg.GetAttachEnable() == 1 {
		errR = server.InsertInteractionAndSend(conn, &msg)
	} else {
		errR = server.DeleteInteractionAndSend(conn, &msg)
	}

	if errR != nil {
		return
	}

	res := mc_metadata.P2C_ReportInteractionAttach{
		UuId:            msg.GetUuId(),
		AttachEnable:    msg.GetAttachEnable(),
		ObjectIndex:     msg.GetObjectIndex(),
		AnimMontageName: msg.GetAnimMontageName(),
		DestinationX:    msg.GetDestinationX(),
		DestinationY:    msg.GetDestinationY(),
		DestinationZ:    msg.GetDestinationZ(),
	}
	log.Println("P2C_ReportInteractionAttach sent from ", res.GetUuId())
	server.makeReportToRegion(conn, &res, constant.P2C_ReportInteractionAttach)
}

func (server *Server) DeleteInteractionAndSend(conn mnet.Client, msg *mc_metadata.C2P_RequestInteractionAttach) error {

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return err
	}

	att := mc_metadata.P2C_ResultInteractionAttach{
		ErrorCode: -1,
	}

	if plr.conn.GetPlayer().Interaction == nil {
		plr.conn.GetPlayer().Interaction = &model.Interaction{}
	}
	plr.conn.GetPlayer().Interaction.ObjectIndex = msg.GetObjectIndex()
	plr.conn.GetPlayer().Interaction.AttachEnabled = msg.GetAttachEnable()
	plr.conn.GetPlayer().Interaction.AnimMontageName = msg.GetAnimMontageName()
	plr.conn.GetPlayer().Interaction.DestinationX = msg.GetDestinationX()
	plr.conn.GetPlayer().Interaction.DestinationY = msg.GetDestinationY()
	plr.conn.GetPlayer().Interaction.DestinationZ = msg.GetDestinationZ()
	server.setPlayer(plr.conn.GetPlayer())

	data, err := proto.MakeResponse(&att, constant.P2C_ResultInteractionAttach)
	if err != nil {
		log.Println("ERROR P2C_ResultInteractionAttach", err)
		return err
	}

	server.sendMsgToMe(data, conn)
	return nil
}

func (server *Server) InsertInteractionAndSend(conn mnet.Client, msg *mc_metadata.C2P_RequestInteractionAttach) error {

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return nil
	}

	att := mc_metadata.P2C_ResultInteractionAttach{
		ErrorCode: -1,
	}

	for i := 0; i < len(server.players); i++ {
		if server.players[i].conn.GetPlayer().Interaction != nil &&
			plr.conn.GetPlayer().UId != server.players[i].conn.GetPlayer().UId &&
			msg.ObjectIndex == server.players[i].conn.GetPlayer().Interaction.ObjectIndex {
			att.ErrorCode = constant.ErrorCodeChairNotEmpty
			break
		}
	}

	if att.ErrorCode == -1 {
		if plr.conn.GetPlayer().Interaction == nil {
			plr.conn.GetPlayer().Interaction = &model.Interaction{}
		}
		plr.conn.GetPlayer().Interaction.ObjectIndex = msg.GetObjectIndex()
		plr.conn.GetPlayer().Interaction.AttachEnabled = msg.GetAttachEnable()
		plr.conn.GetPlayer().Interaction.AnimMontageName = msg.GetAnimMontageName()
		plr.conn.GetPlayer().Interaction.DestinationX = msg.GetDestinationX()
		plr.conn.GetPlayer().Interaction.DestinationY = msg.GetDestinationY()
		plr.conn.GetPlayer().Interaction.DestinationZ = msg.GetDestinationZ()
		server.setPlayer(plr.conn.GetPlayer())
	}

	data, err := proto.MakeResponse(&att, constant.P2C_ResultInteractionAttach)
	if err != nil {
		log.Println("ERROR P2C_ResultInteractionAttach", err)
		return err
	}

	server.sendMsgToMe(data, conn)
	if att.ErrorCode == -1 {
		return nil
	} else {
		return errors.New("chair not empty")
	}
}

func (server *Server) playerPlayAnimation(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestPlayMontage{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ReportPlayMontage{
		UuId:    msg.GetUuId(),
		AnimTid: msg.GetAnimTid(),
	}

	server.makeReportToRegion(conn, &res, constant.P2C_ReportPlayMontage)
}

func (server *Server) playerRegionRoleChecking(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestRoleChecking{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	is := 0
	for i := 0; i < len(server.players); i++ {
		if server.players[i].conn.GetPlayer().Interaction != nil &&
			server.players[i].conn.GetPlayer().Character.Role > 1 {
			is = 1
			break
		}
	}

	res := &mc_metadata.P2C_ResultRoleChecking{
		UuId:      msg.GetUuId(),
		IsTeacher: int32(is),
	}

	data, err := proto.MakeResponse(res, constant.P2C_ResultRoleChecking)
	if err != nil {
		log.Println("ERROR P2C_ResultRoleChecking", msg.GetUuId())
		return
	}
	log.Println("P2C_ResultRoleChecking")
	server.sendMsgToMe(data, conn)
	data = nil
}

func (server *Server) playerEnterToRoom(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMetaSchoolEnter{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plr.conn.GetPlayer().Character.Role = msg.TeacherEnable
	plr.conn.GetPlayer().Interaction = &model.Interaction{
		AttachEnabled: 1,
		ObjectIndex:   0,
	}
	server.setPlayer(plr.conn.GetPlayer())

	reportEnter := &mc_metadata.P2C_ReportMetaSchoolEnter{
		UuId:          msg.GetUuId(),
		TeacherEnable: msg.GetTeacherEnable(),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			UuId:     plr.conn.GetPlayer().UId,
			Nickname: plr.conn.GetPlayer().Character.NickName,
			Role:     plr.conn.GetPlayer().Character.Role,
			Hair:     plr.conn.GetPlayer().Character.Hair,
			Top:      plr.conn.GetPlayer().Character.Top,
			Bottom:   plr.conn.GetPlayer().Character.Bottom,
			Clothes:  plr.conn.GetPlayer().Character.Clothes,
		},
	}

	data1, err2 := proto.MakeResponse(reportEnter, constant.P2C_ReportMetaSchoolEnter)
	if err2 != nil {
		log.Println("ERROR P2C_ResultWhisper", msg.GetUuId())
		return
	}
	log.Println("P2C_ResultMetaSchoolEnter sendMsgToRegion")
	server.sendMsgToRegion(data1, plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)

	res := &mc_metadata.P2C_ResultMetaSchoolEnter{
		UuId:          msg.GetUuId(),
		TeacherEnable: msg.GetTeacherEnable(),
		DataSchool:    proto.ConvertPlayersToRoomReport(server.getRoomPlayers(msg.GetUuId())),
	}

	data2, err3 := proto.MakeResponse(res, constant.P2C_ResultMetaSchoolEnter)
	if err3 != nil {
		log.Println("ERROR P2C_ResultWhisper", msg.GetUuId())
		return
	}

	server.sendMsgToMe(data2, conn)
	data1 = nil
	data2 = nil
}

func (server *Server) playerLeaveFromRoom(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMetaSchoolLeave{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil {
		log.Println("Failed to parse data:", err)
		return
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plr.conn.GetPlayer().Interaction = nil
	server.setPlayer(plr.conn.GetPlayer())

	res := &mc_metadata.P2C_ReportMetaSchoolLeave{
		UuId: msg.GetUuId(),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			UuId:     plr.conn.GetPlayer().UId,
			Nickname: plr.conn.GetPlayer().Character.NickName,
			Role:     plr.conn.GetPlayer().Character.Role,
			Hair:     plr.conn.GetPlayer().Character.Hair,
			Top:      plr.conn.GetPlayer().Character.Top,
			Bottom:   plr.conn.GetPlayer().Character.Bottom,
			Clothes:  plr.conn.GetPlayer().Character.Clothes,
		},
	}

	data, err := proto.MakeResponse(res, constant.P2C_ReportMetaSchoolLeave)
	if err != nil {
		log.Println("ERROR P2C_ReportMetaSchoolLeave", msg.GetUuId())
		return
	}
	server.sendMsgToRegion(data, plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)

	data = nil
}

func (server *Server) playerMovement(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestMove{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetMovementData().GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ReportMove{
		MovementData: proto.MakeMovementData(msg.GetMovementData()),
	}

	server.makeReportToRegion(conn, &res, constant.P2C_ReportMove)
	go server.updateUserLocation(msg.GetMovementData())
}

func (server *Server) playerInfo(conn mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestPlayerInfo{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ResultPlayerInfo{
		ErrorCode: constant.NoError,
	}

	if server.isPlayerOnline(msg.GetNickname(), msg.GetUuId()) {
		res.ErrorCode = constant.ErrorCodeAlreadyOnline

		data, err := proto.MakeResponse(&res, constant.P2C_ResultPlayerInfo)
		if err != nil {
			log.Println("ERROR P2C_ResultPlayerInfo Already Online", msg.GetUuId())
			return
		}
		conn.Send(data)
		return
	}

	plr, err1 := db.GetLoggedDataByName(msg)

	if err1 != nil {
		log.Println("Inserting new user", msg.GetUuId())
		iErr := db.InsertNewAccount(&plr)
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

	data, err := proto.MakeResponse(&res, constant.P2C_ResultPlayerInfo)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", msg.GetUuId())
		return
	}
	conn.Send(data)
	//server.sendMsgToMe(data, conn)
	data = nil
}

func (server *Server) playerLogout(conn mnet.Client, reader mpacket.Reader) {

	msg := mc_metadata.C2P_RequestLogoutUser{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	res := mc_metadata.P2C_ReportLogoutUser{
		UuId: msg.GetUuId(),
	}

	data, err := proto.MakeResponse(&res, constant.P2C_ReportLogoutUser)
	if err != nil {
		log.Println("ERROR P2C_ResultLoginUser", msg.GetUuId())
		return
	}

	server.sendMsgToAll(data, msg.GetUuId())
	server.ClientDisconnected(conn)
	data = nil
}

func (server *Server) chatSendAll(conn mnet.Client, reader mpacket.Reader) {
	msg := mc_metadata.C2P_RequestAllChat{}
	err := proto.Unmarshal(reader.GetBuffer(), &msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := mc_metadata.P2C_ReportAllChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	data, err := proto.MakeResponse(&res, constant.P2C_ReportAllChat)
	if err != nil {
		log.Println("ERROR P2C_ReportAllChat", msg.GetUuId())
		return
	}

	server.sendMsgToAll(data, msg.GetUuId())
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	db.InsertPublicMessage(plr.conn.GetPlayer().CharacterID, constant.World, msg.GetChat())

	toMe := mc_metadata.P2C_ResultAllChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	dataMe, err := proto.MakeResponse(&toMe, constant.P2C_ResultAllChat)
	if err != nil {
		log.Println("ERROR P2C_ResultWhisper", msg.GetUuId())
		return
	}

	server.sendMsgToMe(dataMe, conn)
	dataMe = nil
	data = nil
}

func (server *Server) chatSendRegion(conn mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestRegionChat{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := &mc_metadata.P2C_ReportRegionChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	data, err := proto.MakeResponse(res, constant.P2C_ReportRegionChat)
	if err != nil {
		log.Println("ERROR P2C_ReportAllChat", msg.GetUuId())
		return
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	server.sendMsgToRegion(data, plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)
	db.InsertPublicMessage(
		plr.conn.GetPlayer().CharacterID,
		plr.conn.GetPlayer().RegionID,
		msg.GetChat())

	toMe := &mc_metadata.P2C_ResultRegionChat{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	dataMe, err := proto.MakeResponse(toMe, constant.P2C_ResultRegionChat)
	if err != nil {
		log.Println("ERROR P2C_ResultWhisper", msg.GetUuId())
		return
	}

	server.sendMsgToMe(dataMe, conn)
	dataMe = nil
	data = nil
}

func (server *Server) chatSendWhisper(conn mnet.Client, reader mpacket.Reader) {
	msg := &mc_metadata.C2P_RequestWhisper{}
	err := proto.Unmarshal(reader.GetBuffer(), msg)
	if err != nil || len(msg.GetUuId()) == 0 {
		log.Println("Failed to parse data:", err)
		return
	}

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		log.Println("Player not found", err)
		return
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)

	res := &mc_metadata.P2C_ReportWhisper{
		UuId:     msg.GetUuId(),
		Nickname: msg.GetPlayerNickname(),
		Chat:     msg.GetChat(),
		Time:     t,
	}

	data, err := proto.MakeResponse(res, constant.P2C_ReportWhisper)
	if err != nil {
		log.Println("ERROR P2C_ReportAllChat", msg.GetUuId())
		return
	}

	toMe := &mc_metadata.P2C_ResultWhisper{
		UuId:      msg.GetUuId(),
		Nickname:  msg.GetTargetNickname(),
		Chat:      msg.GetChat(),
		Time:      t,
		ErrorCode: constant.NoError,
	}

	targetPlayer := server.findCharacterIDByNickname(msg.GetTargetNickname())
	db.InsertWhisperMessage(
		plr.conn.GetPlayer().CharacterID,
		targetPlayer.CharacterID,
		msg.GetChat())

	if targetPlayer != nil {
		server.sendMsgToPlayer(data, targetPlayer.UId)
	} else {
		toMe.ErrorCode = constant.ErrorUserOffline
	}

	dataMe, err := proto.MakeResponse(toMe, constant.P2C_ResultWhisper)
	if err != nil {
		log.Println("ERROR P2C_ResultWhisper", msg.GetUuId())
		return
	}

	server.sendMsgToMe(dataMe, conn)
	dataMe = nil
	data = nil
}

func (server *Server) findCharacterIDByUID(uID string) *model.Player {
	for i := 0; i < len(server.players); i++ {
		if uID == server.players[i].conn.GetPlayer().UId {
			return server.players[i].conn.GetPlayer()
		}
	}
	return nil
}

func (server *Server) findCharacterIDByNickname(nickname string) *model.Player {
	for i := 0; i < len(server.players); i++ {
		if nickname == server.players[i].conn.GetPlayer().Character.NickName {
			return server.players[i].conn.GetPlayer()
		}
	}
	return nil
}

func (server *Server) makeReportToRegion(conn mnet.Client, msg proto2.Message, mType uint32) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		log.Println("Player not found", err)
		return
	}

	res, err := proto.MakeResponse(msg, mType)
	if err != nil {
		log.Println("DATA_RESPONSE_MOVEMENT_ERROR", err)
	}

	server.sendMsgToRegion(res, plr.conn.GetPlayer().UId, plr.conn.GetPlayer().RegionID)
	res = nil
}

func (server *Server) updateUserLocation(msg *mc_metadata.Movement) {
	for i := 0; i < len(server.players); i++ {
		if msg.GetUuId() == server.players[i].conn.GetPlayer().UId {
			server.players[i].conn.GetPlayer().Character.PosX = msg.GetDestinationX()
			server.players[i].conn.GetPlayer().Character.PosY = msg.GetDestinationY()
			server.players[i].conn.GetPlayer().Character.PosZ = msg.GetDestinationZ()
			server.players[i].conn.GetPlayer().Character.RotX = msg.GetDeatinationRotationX()
			server.players[i].conn.GetPlayer().Character.RotY = msg.GetDeatinationRotationY()
			server.players[i].conn.GetPlayer().Character.RotZ = msg.GetDeatinationRotationZ()
			break
		}
	}
}

func (server Server) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	packetPlayerEmoticon := func(charID int32, emotion int32) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
		p.WriteInt32(charID)
		p.WriteInt32(emotion)

		return p
	}

	inst.sendExcept(packetPlayerEmoticon(plr.id, emote), plr.conn)
}

func (server *Server) playerUseMysticDoor(conn mnet.Client, reader mpacket.Reader) {
	// doorID := reader.ReadInt32()
	// fromTown := reader.ReadBool()
}

func (server Server) playerAddStatPoint(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if player.ap > 0 {
		player.giveAP(-1)
	}

	statID := reader.ReadInt32()

	switch statID {
	case constant.StrID:
		player.giveStr(1)
	case constant.DexID:
		player.giveDex(1)
	case constant.IntID:
		player.giveInt(1)
	case constant.LukID:
		player.giveLuk(1)
	default:
		fmt.Println("unknown stat id:", statID)
	}
}

func (server *Server) playerRequestAvatarInfoWindow(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	conn.Send(packetPlayerAvatarSummaryWindow(plr.id, *plr))
}

func (server *Server) playerPassiveRegen(conn mnet.Client, reader mpacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if player.hp == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
		return
	}

	if hp > 0 {
		player.giveHP(int16(hp))
	} else if mp > 0 {
		player.giveMP(int16(mp))
	}
}

func (server Server) playerUseChair(conn mnet.Client, reader mpacket.Reader) {
	fmt.Println("use chair:", reader)
	// chairID := reader.ReadInt32()
}

func (server Server) playerStand(conn mnet.Client, reader mpacket.Reader) {
	fmt.Println(reader)
	if reader.ReadInt16() == -1 {

	} else {
	}
}

func (server Server) playerAddSkillPoint(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if plr.sp < 1 {
		return // hacker
	}

	skillID := reader.ReadInt32()
	skill, ok := plr.skills[skillID]

	if ok {
		skill, err = createPlayerSkillFromData(skillID, skill.Level+1)

		if err != nil {
			return
		}

		plr.updateSkill(skill)
	} else {
		// check if class can have skill
		baseSkillID := skillID / 10000
		if !validateSkillWithJob(plr.job, baseSkillID) {
			conn.Send(packetPlayerNoChange())
			return
		}

		skill, err = createPlayerSkillFromData(skillID, 1)

		if err != nil {
			return
		}

		plr.updateSkill(skill)
	}

	plr.giveSP(-1)
}

func validateSkillWithJob(jobID int16, baseSkillID int32) bool {
	if baseSkillID == 0 { // Beginner skills
		return true
	}

	switch jobID {
	case constant.WarriorJobID:
		if baseSkillID != constant.WarriorJobID {
			return false
		}
	case constant.FighterJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.FighterJobID {
			return false
		}
	case constant.CrusaderJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.FighterJobID && baseSkillID != constant.CrusaderJobID {
			return false
		}
	case constant.PageJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.PageJobID {
			return false
		}
	case constant.WhiteKnightJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.PageJobID && baseSkillID != constant.WhiteKnightJobID {
			return false
		}
	case constant.SpearmanJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.SpearmanJobID {
			return false
		}
	case constant.DragonKnightJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.SpearmanJobID && baseSkillID != constant.DragonKnightJobID {
			return false
		}
	case constant.MagicianJobID:
		if baseSkillID != constant.MagicianJobID {
			return false
		}
	case constant.FirePoisonWizardJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.FirePoisonWizardJobID {
			return false
		}
	case constant.FirePoisonMageJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.FirePoisonWizardJobID && baseSkillID != constant.FirePoisonMageJobID {
			return false
		}
	case constant.IceLightWizardJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.IceLightWizardJobID {
			return false
		}
	case constant.IceLightMageJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.IceLightWizardJobID && baseSkillID != constant.IceLightMageJobID {
			return false
		}
	case constant.ClericJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.ClericJobID {
			return false
		}
	case constant.PriestJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.ClericJobID && baseSkillID != constant.PriestJobID {
			return false
		}
	case constant.BowmanJobID:
		if baseSkillID != constant.BowmanJobID {
			return false
		}
	case constant.HunterJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.HunterJobID {
			return false
		}
	case constant.RangerJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.HunterJobID && baseSkillID != constant.RangerJobID {
			return false
		}
	case constant.CrossbowmanJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.CrossbowmanJobID {
			return false
		}
	case constant.SniperJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.CrossbowmanJobID && baseSkillID != constant.SniperJobID {
			return false
		}
	case constant.ThiefJobID:
		if baseSkillID != constant.ThiefJobID {
			return false
		}
	case constant.AssassinJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.AssassinJobID {
			return false
		}
	case constant.HermitJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.AssassinJobID && baseSkillID != constant.HermitJobID {
			return false
		}
	case constant.BanditJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.BanditJobID {
			return false
		}
	case constant.ChiefBanditJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.BanditJobID && baseSkillID != constant.ChiefBanditJobID {
			return false
		}
	case constant.GmJobID:
		if baseSkillID != constant.GmJobID {
			return false
		}
	case constant.SuperGmJobID:
		if baseSkillID != constant.GmJobID && baseSkillID != constant.SuperGmJobID {
			return false
		}
	default:
		return false
	}

	return true
}

func (server Server) playerUsePortal(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if plr.portalCount != reader.ReadByte() {
		conn.Send(packetPlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()
	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	srcInst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	switch entryType {
	case 0:
		if plr.hp == 0 {
			dstField, ok := server.fields[field.Data.ReturnMap]

			if !ok {
				return
			}

			dstInst, err := dstField.getInstance(plr.inst.id)

			if err != nil {
				dstInst, err = dstField.getInstance(0)

				if err != nil {
					return
				}
			}

			portal, err := dstInst.getRandomSpawnPortal()

			if err != nil {
				conn.Send(packetPlayerNoChange())
				return
			}

			server.warpPlayer(plr, dstField, portal)
			plr.setHP(50)
			// TODO: reduce exp
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := srcInst.getPortalFromName(portalName)

		if !plr.checkPos(srcPortal.pos, 100, 100) { // trying to account for lag whilst preventing teleporting
			if conn.GetAdminLevel() > 0 {
				conn.Send(packetMessageRedText("Portal - " + srcPortal.pos.String() + " Player - " + plr.pos.String()))
			}

			conn.Send(packetPlayerNoChange())
			return
		}

		if err != nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		dstField, ok := server.fields[srcPortal.destFieldID]

		if !ok {
			conn.Send(packetPlayerNoChange())
			return
		}

		dstInst, err := dstField.getInstance(plr.inst.id)

		if err != nil {
			if dstInst, err = dstField.getInstance(0); err != nil {
				return
			}
		}

		dstPortal, err := dstInst.getPortalFromName(srcPortal.destName)

		if err != nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		server.warpPlayer(plr, dstField, dstPortal)
	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
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

func (server Server) playerMoveInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	inv := reader.ReadByte()
	pos1 := reader.ReadInt16()
	pos2 := reader.ReadInt16()
	amount := reader.ReadInt16()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	var maxInvSize byte

	switch inv {
	case 1:
		maxInvSize = plr.equipSlotSize
	case 2:
		maxInvSize = plr.useSlotSize
	case 3:
		maxInvSize = plr.setupSlotSize
	case 4:
		maxInvSize = plr.etcSlotSize
	case 5:
		maxInvSize = plr.cashSlotSize
	}

	if pos2 > int16(maxInvSize) {
		return // Moving to item slot the user does not have
	}

	err = plr.moveItem(pos1, pos2, amount, inv)

	if err != nil {
		log.Println(err)
	}
}

func (server Server) playerDropMesos(conn mnet.Client, reader mpacket.Reader) {
	amount := reader.ReadInt32()
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	err = plr.dropMesos(amount)
	if err != nil {
		log.Println(err)
	}

	plr.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, amount, plr.pos, true, plr.id, plr.id)

}

func (server Server) playerUseInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	slot := reader.ReadInt16()
	itemid := reader.ReadInt32()

	item, err := plr.takeItem(itemid, slot, 1, 2)
	if err != nil {
		log.Println(err)
	}
	item.use(plr)

}

func (server Server) playerPickupItem(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	posx := reader.ReadInt16()
	posy := reader.ReadInt16()
	dropID := reader.ReadInt32()

	pos := pos{
		x: posx,
		y: posy,
	}

	err, drop := plr.inst.dropPool.findDropFromID(dropID)

	if err != nil {
		plr.send(packetDropNotAvailable())
		log.Printf("drop Unavailable: %v\nError: %s", drop, err)
		return
	}

	if plr.pos.x-pos.x > 800 || plr.pos.y-pos.y > 600 {
		// Hax
		log.Printf("player: %s tried to pickup an item from far away", plr.name)
		plr.send(packetDropNotAvailable())
		plr.send(packetInventoryDontTake())
		return
	}

	if drop.mesos > 0 {
		plr.giveMesos(drop.mesos)
	} else {
		err = plr.giveItem(drop.item)
		if err != nil {
			plr.send(packetInventoryFull())
			plr.send(packetInventoryDontTake())
			return
		}

	}

	plr.inst.dropPool.playerAttemptPickup(drop, plr)

}

func (server Server) playerTakeDamage(conn mnet.Client, reader mpacket.Reader) {
	// 21 FF  or -1 is mob
	// 21 FE  or -2 is bump
	// Anything bigger than -1 is magic

	dmgType := int8(reader.ReadByte())

	if dmgType >= -1 {
		server.mobDamagePlayer(conn, reader, dmgType)
	} else if dmgType == -2 {
		server.playerBumpDamage(conn, reader)
	} else {
		log.Printf("\nUNKNOWN DAMAGE PACKET: %v", reader.String())
	}
}

func (server Server) playerBumpDamage(conn mnet.Client, reader mpacket.Reader) {
	damage := reader.ReadInt32() // Damage amount

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plr.damagePlayer(int16(damage))

}

func (server Server) getPlayerInstance(conn mnet.Client, reader mpacket.Reader) (*fieldInstance, error) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return nil, err
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return nil, err
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (server *Server) playerBuddyOperation(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1: // Add
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		if plr.buddyListFull() {
			conn.Send(packetBuddyPlayerFullList())
			return
		}

		name := reader.ReadString(reader.ReadInt16())

		var charID int32
		var accountID int32
		var buddyListSize int32

		err = db.Maria.QueryRow("SELECT id,accountID,buddyListSize FROM characters WHERE BINARY name=? and worldID=?", name, conn.GetWorldID()).Scan(&charID, &accountID, &buddyListSize)

		if err != nil || accountID == conn.GetAccountID() {
			conn.Send(packetBuddyNameNotRegistered())
			return
		}

		if plr.hasBuddy(charID) {
			conn.Send(packetBuddyAlreadyAdded())
			return
		}

		var recepientBuddyCount int32
		err = db.Maria.QueryRow("SELECT COUNT(*) FROM buddy WHERE characterID=1 and accepted=1").Scan(&recepientBuddyCount)

		if err != nil {
			log.Fatal(err)
			return
		}

		if recepientBuddyCount >= buddyListSize {
			conn.Send(packetBuddyOtherFullList())
			return
		}

		if conn.GetAdminLevel() == 0 {
			var gm bool
			err = db.Maria.QueryRow("SELECT adminLevel from accounts where accountID=?", accountID).Scan(&gm)

			if err != nil {
				log.Fatal(err)
				return
			}

			if gm {
				conn.Send(packetBuddyIsGM())
				return
			}
		}

		query := "INSERT INTO buddy(characterID,friendID) VALUES(?,?)"

		if _, err = db.Maria.Exec(query, charID, plr.id); err != nil {
			log.Fatal(err)
			return
		}

		if recepient, err := server.players.getFromID(charID); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(1, charID, plr.id, plr.name, server.id))
		} else {
			recepient.send(packetBuddyReceiveRequest(plr.id, plr.name, int32(server.id)))
		}
	case 2: // Accept request
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		friendID := reader.ReadInt32()

		var friendName string
		var friendChannel int32
		var cashShop bool

		err = db.Maria.QueryRow("SELECT name,channelID,inCashShop FROM characters WHERE id=?", friendID).Scan(&friendName, &friendChannel, &cashShop)

		if err != nil {
			log.Fatal(err)
			return
		}

		query := "UPDATE buddy set accepted=1 WHERE characterID=? and friendID=?"

		if _, err := db.Maria.Exec(query, plr.id, friendID); err != nil {
			log.Fatal(err)
			return
		}

		query = "INSERT INTO buddy(characterID,friendID,accepted) VALUES(?,?,?)"

		if _, err := db.Maria.Exec(query, friendID, plr.id, 1); err != nil {
			log.Fatal(err)
			return
		}

		if friendChannel == -1 {
			plr.addOfflineBuddy(friendID, friendName)
		} else {
			plr.addOnlineBuddy(friendID, friendName, friendChannel)
		}

		if recepient, err := server.players.getFromID(friendID); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(2, friendID, plr.id, plr.name, server.id))
		} else {
			// Need to set the buddy to be offline for the logged in message to appear before setting online
			recepient.addOfflineBuddy(plr.id, plr.name)
			recepient.send(packetBuddyOnlineStatus(plr.id, int32(server.id)))
			recepient.addOnlineBuddy(plr.id, plr.name, int32(server.id))
		}
	case 3: // Delete/reject friend
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		id := reader.ReadInt32()

		query := "DELETE FROM buddy WHERE (characterID=? AND friendID=?) OR (characterID=? AND friendID=?)"

		if _, err = db.Maria.Exec(query, id, plr.id, plr.id, id); err != nil {
			log.Fatal(err)
			return
		}

		plr.removeBuddy(id)

		if recepient, err := server.players.getFromID(id); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(3, id, plr.id, "", server.id))
		} else {
			recepient.removeBuddy(plr.id)
		}
	default:
		log.Println("Unknown buddy operation:", op)
	}
}

func (server *Server) playerPartyInfo(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1: // create party
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		if plr.party != nil {
			plr.send(packetPartyAlreadyJoined())
			return
		}

		server.world.Send(internal.PacketChannelPartyCreateRequest(plr.id, server.id, plr.mapID, int32(plr.job), int32(plr.level), plr.name))
	case 2: // leave party
		if b := reader.ReadByte(); b != 0 { // Not sure what this byte/bool does
			log.Println("Leave party byte is not zero:", b)
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			return
		}

		partyID := plr.party.id

		server.world.Send(internal.PacketChannelPartyLeave(partyID, plr.id, plr.party.leader(plr.id)))
	case 3: // accept
		partyID := reader.ReadInt32()

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		server.world.Send(internal.PacketChannelPartyAccept(partyID, plr.id, int32(server.id), plr.mapID, int32(plr.job), int32(plr.level), plr.name))
	case 4: // invite
		id := reader.ReadInt32()

		recipient, err := server.players.getFromID(id)

		if err != nil {
			conn.Send(packetPartyUnableToFindPlayer())
			return
		}

		if recipient.party != nil {
			conn.Send(packetPartyAlreadyJoined())
			return
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			plr.send(packetPartyUnableToFindPlayer())
			return
		}

		if plr.party.full() {
			plr.send(packetPartyToJoinIsFull())
			return
		}

		recipient.send(packetPartyInviteNotice(plr.party.id, plr.name))
	case 5: // expel
		playerID := reader.ReadInt32()

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			plr.send(packetPartyUnableToFindPlayer())
			return
		}

		server.world.Send(internal.PacketChannelPartyExpel(plr.party.id, playerID))
	default:
		log.Println("Unknown party info type:", op, reader)
	}
}

func (server Server) chatGroup(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	op := reader.ReadByte()

	switch op {
	case 0: // buddy
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(1, plr.name, buffer))
	case 1: // party
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(2, plr.name, buffer))
	case 2: // guild
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(3, plr.name, buffer))
	default:
		log.Println("Unknown group chat type:", op, reader)
	}
}

func (server Server) chatSlashCommand(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 5: // find / map button in friend
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}
		name := reader.ReadString(reader.ReadInt16())

		var accountID int32
		var channelID int8
		var mapID int32 = -1
		var inCashShop bool

		err = db.Maria.QueryRow("SELECT accountID,channelID,mapID,inCashShop FROM characters WHERE BINARY name=? AND worldID=?", name, conn.GetWorldID()).Scan(&accountID, &channelID, &mapID, &inCashShop)

		if err != nil || channelID == -1 {
			plr.send(packetMessageFindResult(name, false, false, false, -1))
			return
		}

		var isGM bool

		err = db.Maria.QueryRow("SELECT adminLevel from accounts where accountID=?", accountID).Scan(&isGM)

		if err != nil {
			log.Fatal(err)
			return
		}

		if isGM {
			plr.send(packetMessageFindResult(name, false, inCashShop, false, mapID))
		} else {
			plr.send(packetMessageFindResult(name, true, inCashShop, byte(channelID) == server.id, mapID))
		}
	case 6: // whispher
		recepientName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())

		if receiver, err := server.players.getFromName(recepientName); err != nil {
			var online bool
			err := db.Maria.QueryRow("SELECT COUNT(*) FROM characters WHERE BINARY name=? AND worldID=? AND channelID != -1", recepientName, conn.GetWorldID()).Scan(&online)

			if err != nil || !online {
				conn.Send(packetMessageRedText("Incorrect character name"))
				return
			}

			plr, err := server.players.getFromConn(conn)

			if err != nil {
				return
			}

			plr.send(packetMessageWhisper(plr.name, msg, server.id))
			server.world.Send(internal.PacketChannelWhispherChat(recepientName, plr.name, msg, server.id))
		} else {
			plr, err := server.players.getFromConn(conn)

			if err != nil {
				return
			}

			plr.send(packetMessageWhisper(plr.name, msg, server.id))
			receiver.send(packetMessageWhisper(plr.name, msg, server.id))
		}
	default:
		log.Println("Unkown slash command type:", op, reader)
	}
}

func (server Server) mobControl(conn mnet.Client, reader mpacket.Reader) {
	mobSpawnID := reader.ReadInt32()
	moveID := reader.ReadInt16()
	bits := reader.ReadByte()
	action := reader.ReadInt8()
	skillData := reader.ReadUint32()

	skillPossible := (bits & 0x0F) != 0

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	inst, err := server.getPlayerInstance(conn, reader)
	if err != nil {
		return
	}

	moveData, finalData := parseMovement(reader)

	moveBytes := generateMovementBytes(moveData)

	inst.lifePool.mobAcknowledge(mobSpawnID, plr, moveID, skillPossible, byte(action), skillData, moveData, finalData, moveBytes)

}

func (server Server) mobDamagePlayer(conn mnet.Client, reader mpacket.Reader, mobAttack int8) {
	damage := reader.ReadInt32() // Damage amount
	healSkillID := int32(0)

	if damage < -1 {
		return
	}

	reducedDamage := damage

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]
	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	if err != nil {
		return
	}

	var mob monster
	var mobSkillID, mobSkillLevel byte = 0, 0

	if mobAttack < -1 {
		mobSkillLevel = reader.ReadByte()
		mobSkillID = reader.ReadByte()
	} else {
		magicElement := int32(0)

		if reader.ReadBool() {
			magicElement = reader.ReadInt32()
			_ = magicElement
			// 0 = no element (Grendel the Really Old, 9001001)
			// 1 = Ice (Celion? blue, 5120003)
			// 2 = Lightning (Regular big Sentinel, 3000000)
			// 3 = Fire (Fire sentinel, 5200002)
		}

		spawnID := reader.ReadInt32()
		mobID := reader.ReadInt32()

		mob, err = inst.lifePool.getMobFromID(spawnID)
		if err != nil {
			return
		}

		if mob.id != mobID {
			return
		}

		stance := reader.ReadByte()

		reflected := reader.ReadByte()

		reflectAction := byte(0)
		var reflectX, reflectY int16 = 0, 0

		if reflected > 0 {
			reflectAction = reader.ReadByte()
			reflectX, reflectY = reader.ReadInt16(), reader.ReadInt16()
		}

		// Magic guard dmg absorption

		// Fighter / Page power guard

		// Meso guard

		if !plr.admin() {
			plr.damagePlayer(int16(damage))
		}

		inst.send(packetPlayerReceivedDmg(plr.id, mobAttack, damage, reducedDamage, spawnID, mobID, healSkillID, stance, reflectAction, reflected, reflectX, reflectY))
	}
	if mobSkillID != 0 && mobSkillLevel != 0 {
		// new skill
	}

}

func (server Server) mobDistance(conn mnet.Client, reader mpacket.Reader) {
	/*
		id := reader.ReadInt32()
		distance := reader.ReadInt32()

		Unknown what this packet is for
	*/

}

func (server Server) playerMeleeSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackMelee)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	// if player in party extract

	packetSkillMelee := func(char player, ad attackData) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerUseMeleeSkill)
		p.WriteInt32(char.id)
		p.WriteByte(ad.targets*0x10 + ad.hits)
		p.WriteByte(ad.skillLevel)

		if ad.skillLevel != 0 {
			p.WriteInt32(ad.skillID)
		}

		if ad.facesLeft {
			p.WriteByte(ad.action | (1 << 7))
		} else {
			p.WriteByte(ad.action | 0)
		}

		p.WriteByte(ad.attackType)

		p.WriteByte(char.skills[ad.skillID].Mastery)
		p.WriteInt32(ad.projectileID)

		for _, info := range ad.attackInfo {
			p.WriteInt32(info.spawnID)
			p.WriteByte(info.hitAction)

			if ad.isMesoExplosion {
				p.WriteByte(byte(len(info.damages)))
			}

			for _, dmg := range info.damages {
				p.WriteInt32(dmg)
			}
		}

		return p
	}

	inst.sendExcept(packetSkillMelee(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

func (server Server) playerRangedSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackRanged)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	err = plr.useSkill(data.skillID, data.skillLevel)
	if err != nil {
		// send packet to stop?
		return
	}

	// if player in party extract

	packetSkillRanged := func(char player, ad attackData) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerUseRangedSkill)
		p.WriteInt32(char.id)
		p.WriteByte(ad.targets*0x10 + ad.hits)
		p.WriteByte(ad.skillLevel)

		if ad.skillLevel != 0 {
			p.WriteInt32(ad.skillID)
		}

		if ad.facesLeft {
			p.WriteByte(ad.action | (1 << 7))
		} else {
			p.WriteByte(ad.action | 0)
		}

		p.WriteByte(ad.attackType)

		p.WriteByte(char.skills[ad.skillID].Mastery)
		p.WriteInt32(ad.projectileID)

		for _, info := range ad.attackInfo {
			p.WriteInt32(info.spawnID)
			p.WriteByte(info.hitAction)

			if ad.isMesoExplosion {
				p.WriteByte(byte(len(info.damages)))
			}

			for _, dmg := range info.damages {
				p.WriteInt32(dmg)
			}
		}

		return p
	}

	inst.sendExcept(packetSkillRanged(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

func (server Server) playerMagicSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackMagic)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	err = plr.useSkill(data.skillID, data.skillLevel)
	if err != nil {
		// send packet to stop?
		return
	}

	// if player in party extract

	packetSkillMagic := func(char player, ad attackData) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerUseMagicSkill)
		p.WriteInt32(char.id)
		p.WriteByte(ad.targets*0x10 + ad.hits)
		p.WriteByte(ad.skillLevel)

		if ad.skillLevel != 0 {
			p.WriteInt32(ad.skillID)
		}

		if ad.facesLeft {
			p.WriteByte(ad.action | (1 << 7))
		} else {
			p.WriteByte(ad.action | 0)
		}

		p.WriteByte(ad.attackType)

		p.WriteByte(char.skills[ad.skillID].Mastery)
		p.WriteInt32(ad.projectileID)

		for _, info := range ad.attackInfo {
			p.WriteInt32(info.spawnID)
			p.WriteByte(info.hitAction)

			if ad.isMesoExplosion {
				p.WriteByte(byte(len(info.damages)))
			}

			for _, dmg := range info.damages {
				p.WriteInt32(dmg)
			}
		}

		return p
	}

	inst.sendExcept(packetSkillMagic(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

// Following logic lifted from WvsGlobal
const (
	attackMelee = iota
	attackRanged
	attackMagic
	attackSummon
)

type attackInfo struct {
	spawnID                                                int32
	hitAction, foreAction, frameIndex, calcDamageStatIndex byte
	facesLeft                                              bool
	hitPosition, previousMobPosition                       pos
	hitDelay                                               int16
	damages                                                []int32
}

type attackData struct {
	skillID, summonType, totalDamage, projectileID int32
	isMesoExplosion, facesLeft                     bool
	option, action, attackType                     byte
	targets, hits, skillLevel                      byte

	attackInfo []attackInfo
	playerPos  pos
}

func getAttackInfo(reader mpacket.Reader, player player, attackType int) (attackData, bool) {
	data := attackData{}

	if player.hp == 0 {
		return data, false
	}

	// speed hack check
	if false && (reader.Time-player.lastAttackPacketTime < 350) {
		return data, false
	}

	player.lastAttackPacketTime = reader.Time

	if attackType != attackSummon {
		tByte := reader.ReadByte()
		skillID := reader.ReadInt32()

		if _, ok := player.skills[skillID]; !ok && skillID != 0 {
			return data, false
		}

		data.skillID = skillID

		if data.skillID != 0 {
			data.skillLevel = player.skills[skillID].Level
		}

		// if meso explosion data.IsMesoExplosion = true

		data.targets = tByte / 0x10
		data.hits = tByte % 0x10
		data.option = reader.ReadByte()

		tmp := reader.ReadByte()

		data.action = tmp & 0x7F
		data.facesLeft = (tmp >> 7) == 1
		data.attackType = reader.ReadByte()
	} else {
		data.summonType = reader.ReadInt32()
		data.attackType = reader.ReadByte()
		data.targets = 1
		data.hits = 1
	}

	reader.Skip(4) //checksum info?

	if attackType == attackRanged {
		projectileSlot := reader.ReadInt16() // star/arrow slot
		if projectileSlot == 0 {
			// if soul arrow is not set check for hacks
		} else {
			data.projectileID = -1

			for _, item := range player.use {
				if item.slotID == projectileSlot {
					data.projectileID = item.id
				}
			}
		}
		reader.ReadByte() // ?
		reader.ReadByte() // ?
		reader.ReadByte() // ?
	}

	data.attackInfo = make([]attackInfo, data.targets)

	for i := byte(0); i < data.targets; i++ {
		attack := attackInfo{}
		attack.spawnID = reader.ReadInt32()
		attack.hitAction = reader.ReadByte()

		tmp := reader.ReadByte()
		attack.foreAction = tmp & 0x7F
		attack.facesLeft = (tmp >> 7) == 1
		attack.frameIndex = reader.ReadByte()

		if !data.isMesoExplosion {
			attack.calcDamageStatIndex = reader.ReadByte()
		}

		attack.hitPosition.x = reader.ReadInt16()
		attack.hitPosition.y = reader.ReadInt16()

		attack.previousMobPosition.x = reader.ReadInt16()
		attack.previousMobPosition.y = reader.ReadInt16()

		if attackType == attackSummon {
			reader.Skip(1)
		}

		if data.isMesoExplosion {
			data.hits = reader.ReadByte()
		} else if attackType != attackSummon {
			attack.hitDelay = reader.ReadInt16()
		}

		attack.damages = make([]int32, data.hits)

		for j := byte(0); j < data.hits; j++ {
			dmg := reader.ReadInt32()
			data.totalDamage += dmg
			attack.damages[j] = dmg
		}
		data.attackInfo[i] = attack
	}

	data.playerPos.x = reader.ReadInt16()
	data.playerPos.y = reader.ReadInt16()

	return data, true
}

func (server *Server) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	inst.lifePool.npcAcknowledge(id, plr, data)
}

func (server *Server) npcChatStart(conn mnet.Client, reader mpacket.Reader) {
	npcSpawnID := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	npcData, err := inst.lifePool.getNPCFromSpawnID(npcSpawnID)

	if err != nil {
		return
	}

	// Start npc session
	var controller *npcScriptController

	if program, ok := server.npcScriptStore.scripts[strconv.Itoa(int(npcData.id))]; ok {
		controller, err = createNewnpcScriptController(npcData.id, conn, program, server.warpPlayer, server.fields)
	} else {
		if program, ok := server.npcScriptStore.scripts["default"]; ok {
			controller, err = createNewnpcScriptController(npcData.id, conn, program, server.warpPlayer, server.fields)
		}
	}

	if controller == nil {
		log.Println("Unable to find npc script for:", npcData.id, ".... default.js not found")
		return
	}

	if err != nil {
		log.Println("script init:", err)
	}

	server.npcChat[conn] = controller
	if controller.run(plr) {
		delete(server.npcChat, conn)
	}
}

func (server *Server) npcChatContinue(conn mnet.Client, reader mpacket.Reader) {
	if _, ok := server.npcChat[conn]; !ok {
		return
	}

	controller := server.npcChat[conn]
	controller.state.ClearFlags()

	terminate := false

	msgType := reader.ReadByte()

	switch msgType {
	case 0: // next/back
		value := reader.ReadByte()

		switch value {
		case 0: // back
			controller.state.SetNextBack(false, true)
		case 1: // next
			controller.state.SetNextBack(true, false)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			terminate = true
			log.Println("unknown next/back:", value)
		}
	case 1: // yes/no, ok
		value := reader.ReadByte()

		switch value {
		case 0: // no
			controller.state.SetYesNo(false, true)
		case 1: // yes, ok
			controller.state.SetYesNo(true, false)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			log.Println("unknown yes/no:", value)
		}
	case 2: // string input
		if reader.ReadBool() {
			controller.state.SetTextInput(reader.ReadString(reader.ReadInt16()))
		} else {
			terminate = true
		}
	case 3: // number input
		if reader.ReadBool() {
			controller.state.SetNumberInput(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 4: // select option
		if reader.ReadBool() {
			controller.state.SetOptionSelect(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 5: // style window (no way to discern between cancel button and end chat selection)
		if reader.ReadBool() {
			controller.state.SetOptionSelect(int32(reader.ReadByte()))
		} else {
			terminate = true
		}
	case 6:
		fmt.Println("pet window:", reader)
	default:
		log.Println("Unkown npc chat continue packet:", reader)
	}

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		delete(server.npcChat, conn)
		return
	}

	if terminate || controller.run(plr) {
		delete(server.npcChat, conn)
	}
}

func (server *Server) npcShop(conn mnet.Client, reader mpacket.Reader) {
	getInventoryID := func(id int32) byte {
		return byte(id / 1000000)
	}

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	operation := reader.ReadByte()
	switch operation {
	case 0: // buy
		index := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()

		newItem, err := createAverageItemFromID(itemID, amount)

		if err != nil {
			return
		}

		if controller, ok := server.npcChat[conn]; ok {
			goods := controller.state.Goods()

			if int(index) < len(goods) && index > -1 {
				if len(goods[index]) == 1 { // Default price
					item, err := nx.GetItem(itemID)

					if err != nil {
						return
					}

					plr.giveMesos(-1 * item.Price)
				} else if len(goods[index]) == 2 { // Custom price
					plr.giveMesos(-1 * goods[index][1])
				} else {
					return // bad shop slice
				}

				plr.giveItem(newItem)
				plr.send(packetNpcShopContinue()) //check if needed
			}

		}
	case 1: // sell
		slotPos := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()

		fmt.Println("Selling:", itemID, "[", slotPos, "], amount:", amount)

		item, err := nx.GetItem(itemID)

		if err != nil {
			return
		}

		invID := getInventoryID(itemID)

		plr.takeItem(itemID, slotPos, amount, invID)

		plr.giveMesos(item.Price)
		plr.send(packetNpcShopContinue()) // check if needed
	case 3: // exit
		if _, ok := server.npcChat[conn]; ok {
			delete(server.npcChat, conn) // delete here as we need access to shop goods
		}
	default:
		log.Println("Unkown shop operation packet:", reader)
	}
}

func (server Server) roomWindow(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	pool := inst.roomPool

	if err != nil {
		return
	}

	operation := reader.ReadByte()

	switch operation {
	case roomCreate:
		switch roomType := reader.ReadByte(); roomType {
		case roomTypeOmok:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := newOmokRoom(inst.nextID(), name, password, boardType).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				pool.addRoom(r)
			}
		case roomTypeMemory:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := newMemoryRoom(inst.nextID(), name, password, boardType).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				pool.addRoom(r)
			}
		case roomTypeTrade:
			r, valid := newTradeRoom(inst.nextID()).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				pool.addRoom(r)
			}
		case roomTypePersonalShop:
			log.Println("Personal shop not implemented")
		default:
			log.Println("Unknown room type", roomType)
		}
	case roomSendInvite:
		id := reader.ReadInt32()

		plr2, err := inst.getPlayerFromID(id)

		if err != nil {
			plr.send(packetRoomTradeRequireSameMap())
			return
		}

		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if trade, valid := r.(*tradeRoom); valid {
			trade.sendInvite(plr2)
		}
	case roomReject:
		id := reader.ReadInt32()
		code := reader.ReadByte()

		r, err := pool.getRoom(id)

		if err != nil {
			return
		}

		if trade, valid := r.(*tradeRoom); valid {
			trade.reject(code, plr.name)
		}
	case roomAccept:
		id := reader.ReadInt32()

		r, err := pool.getRoom(id)

		if err != nil {
			plr.send(packetRoomTradeRequireSameMap())
			return
		}

		if reader.ReadBool() {
			password := reader.ReadString(reader.ReadInt16())

			if game, valid := r.(gameRoomer); valid {
				if !game.checkPassword(password, plr) {
					return
				}
			}
		}

		r.addPlayer(plr)

		if _, valid := r.(gameRoomer); valid {
			pool.updateGameBox(r)
		}
	case roomChat:
		msg := reader.ReadString(reader.ReadInt16())

		if len(msg) > 0 {
			r, err := pool.getPlayerRoom(plr.id)

			if err != nil {
				return
			}

			r.chatMsg(plr, msg)
		}
	case roomCloseWindow:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.kickPlayer(plr, 0x0)

			if r.closed() {
				pool.removeRoom(r.id())
			} else {
				pool.updateGameBox(r)
			}
		} else if trade, valid := r.(*tradeRoom); valid {
			trade.removePlayer(plr)
			pool.removeRoom(trade.roomID)
		}
	case roomInsertItem:
		// invTab := reader.ReadByte()
		// itemSlot := reader.ReadInt16()
		// quantity := reader.ReadInt16()
		// tradeWindowSlot := reader.ReadByte()
	case roomMesos:
		// amount := reader.ReadInt32()
	case roomAcceptTrade:
	case roomRequestTie:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestTie(plr)
		}
	case roomRequestTieResult:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			tie := reader.ReadBool()
			game.requestTieResult(tie, plr)

			if r.closed() {
				pool.removeRoom(r.id())
			} else {
				pool.updateGameBox(r)
			}
		}
	case roomForfeit:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.forfeit(plr)

			if r.closed() {
				pool.removeRoom(r.id())
			} else {
				pool.updateGameBox(r)
			}
		}
	case roomRequestUndo:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			game.requestUndo(plr)
		}
	case roomRequestUndoResult:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			undo := reader.ReadBool()
			game.requestUndoResult(undo, plr)
		}
	case roomRequestExitDuringGame:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestExit(true, plr)
		}
	case roomUndoRequestExit:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestExit(false, plr)
		}
	case roomReadyButtonPressed:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.ready(plr)
		}
	case roomUnready:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.unready(plr)
		}
	case roomOwnerExpells:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.expel()
			pool.updateGameBox(r)
		}
	case roomGameStart:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.start()
			pool.updateGameBox(r)
		}
	case roomChangeTurn:
		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.changeTurn()
		}
	case roomPlacePiece:
		x := reader.ReadInt32()
		y := reader.ReadInt32()
		piece := reader.ReadByte()

		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			if game.placePiece(x, y, piece, plr) {
				pool.updateGameBox(r)
			}

			if r.closed() {
				pool.removeRoom(game.roomID)
			}
		}
	case roomSelectCard:
		turn := reader.ReadByte()
		cardID := reader.ReadByte()

		r, err := pool.getPlayerRoom(plr.id)

		if err != nil {
			return
		}

		if game, valid := r.(*memoryRoom); valid {
			if game.selectCard(turn, cardID, plr) {
				pool.updateGameBox(r)
			}

			if r.closed() {
				pool.removeRoom(game.roomID)
			}
		}
	default:
		log.Println("Unknown room operation", operation)
	}
}

// HandleServerPacket from world
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
