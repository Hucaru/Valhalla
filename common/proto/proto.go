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

func AccountReport(uID string, acc model.Character) *mc_metadata.P2C_ReportLoginUser {
	res := &mc_metadata.P2C_ReportLoginUser{
		UuId: uID,
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

func ChannelChangeForNewReport(plr *model.Player) *mc_metadata.P2C_ReportRegionChange {
	res := &mc_metadata.P2C_ReportRegionChange{
		UuId:     plr.UId,
		RegionId: int32(plr.RegionID),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: plr.Character.NickName,
			Hair:     plr.Character.Hair,
			Top:      plr.Character.Top,
			Bottom:   plr.Character.Bottom,
			Clothes:  plr.Character.Clothes,
			Role:     plr.Character.Role,
		},
		SpawnPosX: plr.Character.PosX,
		SpawnPosY: plr.Character.PosY,
		SpawnPosZ: plr.Character.PosZ,
		SpawnRotX: plr.Character.RotX,
		SpawnRotY: plr.Character.RotY,
		SpawnRotZ: plr.Character.RotZ,
	}

	return res
}

func ChannelChangeForOldReport(uID string, acc *model.Character) *mc_metadata.P2C_ReportRegionLeave {
	res := &mc_metadata.P2C_ReportRegionLeave{
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			UuId:     uID,
			Role:     acc.Role,
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
			UuId:     player.UId,
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

func ConvertPlayersToRoomReport(plrs []*model.Player) []*mc_metadata.DataSchool {
	res := make([]*mc_metadata.DataSchool, 0)

	for i := 0; i < len(plrs); i++ {
		interaction := &mc_metadata.P2C_ReportInteractionAttach{}

		if plrs[i].Interaction.IsInteraction {
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
				UuId:     plrs[i].UId,
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
