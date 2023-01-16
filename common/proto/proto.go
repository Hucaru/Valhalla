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

func AccountReport(uID string, acc model.Character) mc_metadata.P2C_ReportLoginUser {
	res := mc_metadata.P2C_ReportLoginUser{
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
			Nickname: plr.GetCharacter().NickName,
			Hair:     plr.GetCharacter().Hair,
			Top:      plr.GetCharacter().Top,
			Bottom:   plr.GetCharacter().Bottom,
			Clothes:  plr.GetCharacter().Clothes,
			Role:     plr.GetCharacter().Role,
		},
		SpawnPosX: plr.GetCharacter().PosX,
		SpawnPosY: plr.GetCharacter().PosY,
		SpawnPosZ: plr.GetCharacter().PosZ,
		SpawnRotX: plr.GetCharacter().RotX,
		SpawnRotY: plr.GetCharacter().RotY,
		SpawnRotZ: plr.GetCharacter().RotZ,
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

func AccountResult(player *model.Player) mc_metadata.P2C_ResultLoginUser {
	PlayerInfo := mc_metadata.P2C_PlayerInfo{
		UuId:     player.UId,
		Nickname: player.GetCharacter().NickName,
		Hair:     player.GetCharacter().Hair,
		Top:      player.GetCharacter().Top,
		Bottom:   player.GetCharacter().Bottom,
		Clothes:  player.GetCharacter().Clothes,
	}

	return mc_metadata.P2C_ResultLoginUser{
		UuId:        player.UId,
		RegionId:    int32(player.RegionID),
		PlayerInfo:  &PlayerInfo,
		SpawnPosX:   player.GetCharacter().PosX,
		SpawnPosY:   player.GetCharacter().PosY,
		SpawnPosZ:   player.GetCharacter().PosZ,
		SpawnRotX:   player.GetCharacter().RotX,
		SpawnRotY:   player.GetCharacter().RotY,
		SpawnRotZ:   player.GetCharacter().RotZ,
		LoggedUsers: []*mc_metadata.P2C_ReportLoginUser{},
	}
}

func RegionResult(player *model.Player) *mc_metadata.P2C_ResultRegionChange {
	return &mc_metadata.P2C_ResultRegionChange{
		UuId:     player.UId,
		RegionId: int32(player.RegionID),
		PlayerInfo: &mc_metadata.P2C_PlayerInfo{
			Nickname: player.GetCharacter().NickName,
			Hair:     player.GetCharacter().Hair,
			Top:      player.GetCharacter().Top,
			Bottom:   player.GetCharacter().Bottom,
			Clothes:  player.GetCharacter().Clothes,
		},
		SpawnPosX:   player.GetCharacter().PosX,
		SpawnPosY:   player.GetCharacter().PosY,
		SpawnPosZ:   player.GetCharacter().PosZ,
		SpawnRotX:   player.GetCharacter().RotX,
		SpawnRotY:   player.GetCharacter().RotY,
		SpawnRotZ:   player.GetCharacter().RotZ,
		RegionUsers: []*mc_metadata.P2C_ReportRegionChange{},
	}
}

func ConvertPlayersToRoomReport(plrs []*model.Player) []*mc_metadata.DataSchool {
	res := make([]*mc_metadata.DataSchool, 0)

	for i := 0; i < len(plrs); i++ {
		interaction := &mc_metadata.P2C_ReportInteractionAttach{}

		if plrs[i].GetInteraction().IsInteraction {
			interaction.ObjectIndex = plrs[i].GetInteraction().ObjectIndex
			interaction.AttachEnable = plrs[i].GetInteraction().AttachEnabled
			interaction.AnimMontageName = plrs[i].GetInteraction().AnimMontageName
			interaction.UuId = plrs[i].UId
			interaction.DestinationX = plrs[i].GetInteraction().DestinationX
			interaction.DestinationY = plrs[i].GetInteraction().DestinationY
			interaction.DestinationZ = plrs[i].GetInteraction().DestinationZ
		}

		res = append(res, &mc_metadata.DataSchool{
			UuId:            plrs[i].UId,
			InteractionData: interaction,
			PlayerInfo: &mc_metadata.P2C_PlayerInfo{
				UuId:     plrs[i].UId,
				Nickname: plrs[i].GetCharacter().NickName,
				Hair:     plrs[i].GetCharacter().Hair,
				Top:      plrs[i].GetCharacter().Top,
				Bottom:   plrs[i].GetCharacter().Bottom,
				Clothes:  plrs[i].GetCharacter().Clothes,
				Role:     plrs[i].GetCharacter().Role,
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
