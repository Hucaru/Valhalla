package proto

import (
	"encoding/binary"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	"google.golang.org/protobuf/proto"
	"log"
)

func GetRequestLoginUser(buff []byte) (*mc_metadata.C2P_RequestLoginUser, error) {
	msg := &mc_metadata.C2P_RequestLoginUser{}
	if err := proto.Unmarshal(buff, msg); err != nil || len(msg.UuId) == 0 {
		log.Fatalln("Failed to parse data:", err)
		return nil, err
	}
	return msg, nil
}

func Unmarshal(buff []byte, msg proto.Message) error {
	return proto.Unmarshal(buff, msg)
}

func AccountReport(uID *string, acc *model.Character) *mc_metadata.P2C_ReportLoginUser {
	res := &mc_metadata.P2C_ReportLoginUser{
		UuId: *uID,
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: acc.NickName,
			Hair:     acc.Hair,
			Top:      acc.Top,
			Bottom:   acc.Bottom,
			Clothes:  acc.Clothes,
		},
		SpawnPosX: acc.PosX,
		SpawnPosY: acc.PosY,
		SpawnPosZ: acc.PosZ,
		SpawnRotX: acc.RotX,
		SpawnRotY: acc.RotY,
		SpawnRotZ: acc.RotZ,
	}

	return res
}

func ChannelChangeForNewReport(uID *string, acc *model.Character) *mc_metadata.P2C_ReportRegionChange {
	res := &mc_metadata.P2C_ReportRegionChange{
		UuId: *uID,
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: acc.NickName,
			Hair:     acc.Hair,
			Top:      acc.Top,
			Bottom:   acc.Bottom,
			Clothes:  acc.Clothes,
		},
		SpawnPosX: acc.PosX,
		SpawnPosY: acc.PosY,
		SpawnPosZ: acc.PosZ,
		SpawnRotX: acc.RotX,
		SpawnRotY: acc.RotY,
		SpawnRotZ: acc.RotZ,
	}

	return res
}

func ChannelChangeForOldReport(acc *model.Character) *mc_metadata.P2C_ReportRegionLeave {
	res := &mc_metadata.P2C_ReportRegionLeave{
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: acc.NickName,
			Hair:     acc.Hair,
			Top:      acc.Top,
			Bottom:   acc.Bottom,
			Clothes:  acc.Clothes,
		},
	}

	return res
}

func AccountResult(player *model.Player) *mc_metadata.P2C_ResultLoginUser {
	return &mc_metadata.P2C_ResultLoginUser{
		UuId:     player.UId,
		RegionId: int32(player.RegionID),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: player.Character.NickName,
			Hair:     player.Character.Hair,
			Top:      player.Character.Top,
			Bottom:   player.Character.Bottom,
			Clothes:  player.Character.Clothes,
		},
		SpawnPosX:   player.Character.PosX,
		SpawnPosY:   player.Character.PosY,
		SpawnPosZ:   player.Character.PosZ,
		SpawnRotX:   player.Character.RotX,
		SpawnRotY:   player.Character.RotY,
		SpawnRotZ:   player.Character.RotZ,
		LoggedUsers: []*mc_metadata.P2C_ReportLoginUser{},
	}
}

func RegionResult(player *model.Player) *mc_metadata.P2C_ResultRegionChange {
	return &mc_metadata.P2C_ResultRegionChange{
		UuId:     player.UId,
		RegionId: int32(player.RegionID),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: player.Character.NickName,
			Hair:     player.Character.Hair,
			Top:      player.Character.Top,
			Bottom:   player.Character.Bottom,
			Clothes:  player.Character.Clothes,
		},
		SpawnPosX:   player.Character.PosX,
		SpawnPosY:   player.Character.PosY,
		SpawnPosZ:   player.Character.PosZ,
		SpawnRotX:   player.Character.RotX,
		SpawnRotY:   player.Character.RotY,
		SpawnRotZ:   player.Character.RotZ,
		RegionUsers: []*mc_metadata.P2C_ReportRegionChange{},
	}
}

func ConvertPlayersToLoginResult(plrs []*model.Player) []*mc_metadata.P2C_ReportLoginUser {
	res := make([]*mc_metadata.P2C_ReportLoginUser, 0)

	for i := 0; i < len(plrs); i++ {
		res = append(res, &mc_metadata.P2C_ReportLoginUser{
			UuId: plrs[i].UId,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: plrs[i].Character.NickName,
				Hair:     plrs[i].Character.Hair,
				Top:      plrs[i].Character.Top,
				Bottom:   plrs[i].Character.Bottom,
				Clothes:  plrs[i].Character.Clothes,
			},
			SpawnPosX: plrs[i].Character.PosX,
			SpawnPosY: plrs[i].Character.PosY,
			SpawnPosZ: plrs[i].Character.PosZ,
			SpawnRotX: plrs[i].Character.RotX,
			SpawnRotY: plrs[i].Character.RotY,
			SpawnRotZ: plrs[i].Character.RotZ,
		})
	}
	return res
}

func ConvertPlayersToRegionReport(plrs []*model.Player) []*mc_metadata.P2C_ReportRegionChange {
	res := make([]*mc_metadata.P2C_ReportRegionChange, 0)

	for i := 0; i < len(plrs); i++ {
		res = append(res, &mc_metadata.P2C_ReportRegionChange{
			UuId: plrs[i].UId,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: plrs[i].Character.NickName,
				Hair:     plrs[i].Character.Hair,
				Top:      plrs[i].Character.Top,
				Bottom:   plrs[i].Character.Bottom,
				Clothes:  plrs[i].Character.Clothes,
			},
			SpawnPosX: plrs[i].Character.PosX,
			SpawnPosY: plrs[i].Character.PosY,
			SpawnPosZ: plrs[i].Character.PosZ,
			SpawnRotX: plrs[i].Character.RotX,
			SpawnRotY: plrs[i].Character.RotY,
			SpawnRotZ: plrs[i].Character.RotZ,
		})
	}
	return res
}

func ConvertPlayersToRoomReport(plrs []*model.Player) []*mc_metadata.DataSchool {
	res := make([]*mc_metadata.DataSchool, 0)

	for i := 0; i < len(plrs); i++ {
		interaction := &mc_metadata.P2C_ReportInteractionAttach{}

		if plrs[i].Interaction != nil {
			interaction.ObjectIndex = plrs[i].Interaction.ObjectIndex
			interaction.AttachEnable = plrs[i].Interaction.AttachEnabled
			interaction.AnimMontageName = plrs[i].Interaction.AnimMontageName
			interaction.UuId = plrs[i].UId
			interaction.DestinationX = plrs[i].Interaction.DestinationX
			interaction.DestinationY = plrs[i].Interaction.DestinationY
			interaction.DestinationZ = plrs[i].Interaction.DestinationZ
		}

		res = append(res, &mc_metadata.DataSchool{
			UuId:            plrs[i].UId,
			InteractionData: interaction,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				Nickname: plrs[i].Character.NickName,
				Hair:     plrs[i].Character.Hair,
				Top:      plrs[i].Character.Top,
				Bottom:   plrs[i].Character.Bottom,
				Clothes:  plrs[i].Character.Clothes,
				Role:     plrs[i].Character.Role,
			},
		})
	}
	return res
}

func MakeMovementData(msg *mc_metadata.Movement) *mc_metadata.Movement {
	return &mc_metadata.Movement{
		UuId:                 msg.GetUuId(),
		DestinationX:         msg.GetDestinationX(),
		DestinationY:         msg.GetDestinationY(),
		DestinationZ:         msg.GetDestinationZ(),
		DeatinationRotationX: msg.GetDeatinationRotationX(),
		DeatinationRotationY: msg.GetDeatinationRotationY(),
		DeatinationRotationZ: msg.GetDeatinationRotationZ(),
		InterpTime:           msg.GetInterpTime(),
	}
}

func ErrorLoginResponse(_err string, uID string) ([]byte, error) {
	return MakeResponse(&mc_metadata.P2C_ResultLoginUserError{
		UuId:  uID,
		Error: _err,
	}, constant.P2C_ResultLoginUserError)
}

func MakeResponse(msg proto.Message, msgType uint32) ([]byte, error) {
	out, err := proto.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal object:", err)
		return nil, err
	}

	result := make([]byte, 0)

	h := make([]byte, 0)
	h = append(h, binary.BigEndian.AppendUint32(h, uint32(len(out)))...)
	h = binary.BigEndian.AppendUint32(h, msgType)
	result = append(result, h...)
	result = append(result, out...)

	msg = nil
	return result, nil
}
