package packet

import (
	"github.com/Hucaru/Valhalla/mnet"
)

type posInter interface {
	SetX(int16)
	GetX() int16
	SetY(int16)
	GetY() int16
}

type lifeInter interface {
	SetID(int32)
	GetID() int32
	SetSpawnID(int32)
	GetSpawnID() int32
	posInter

	SetFoothold(int16)
	GetFoothold() int16
	SetFace(byte)
	GetFace() byte
	GetState() byte
	SetState(byte)
}

type npcInter interface {
	lifeInter
	SetRx0(int16)
	GetRx0() int16
	SetRx1(int16)
	GetRx1() int16
}

type mobInter interface {
	npcInter
	GetFlySpeed() int32
	GetSummoner() mnet.MConnChannel
}
