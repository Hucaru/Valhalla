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

func AccountResponseToAll(acc *model.Account, msgType uint32) ([]byte, error) {
	res, err := MakeResponse(&mc_metadata.P2C_ReportLoginUser{
		UuId:      acc.UId,
		SpawnPosX: acc.PosX,
		SpawnPosY: acc.PosY,
		SpawnPosZ: acc.PosZ,
		SpawnRotX: acc.RotX,
		SpawnRotY: acc.RotY,
		SpawnRotZ: acc.RotZ,
	}, msgType)

	acc = nil
	return res, err
}

func GetResultUser(acc *model.Account) *mc_metadata.P2C_ResultLoginUser {
	return &mc_metadata.P2C_ResultLoginUser{
		UuId:        acc.UId,
		SpawnPosX:   acc.PosX,
		SpawnPosY:   acc.PosY,
		SpawnPosZ:   acc.PosZ,
		SpawnRotX:   acc.RotX,
		SpawnRotY:   acc.RotY,
		SpawnRotZ:   acc.RotZ,
		LoggedUsers: []*mc_metadata.P2C_ReportLoginUser{},
	}
}

func GetLoggedUsers(accounts []*model.Account) []*mc_metadata.P2C_ReportLoginUser {
	res := make([]*mc_metadata.P2C_ReportLoginUser, 0)
	for i := 0; i < len(accounts); i++ {
		res = append(res, &mc_metadata.P2C_ReportLoginUser{
			UuId:      accounts[i].UId,
			SpawnPosX: accounts[i].PosX,
			SpawnPosY: accounts[i].PosY,
			SpawnPosZ: accounts[i].PosZ,
			SpawnRotX: accounts[i].RotX,
			SpawnRotY: accounts[i].RotY,
			SpawnRotZ: accounts[i].RotZ,
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
