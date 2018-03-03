package interfaces

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/gopacket"
)

type OcClientConn interface {
	GetUserID() uint32
}

type ClientConn interface {
	Close()
	Write(gopacket.Packet) error
	String() string
	SetUserID(uint32)
	OcClientConn
	SetAdmin(val bool)
	IsAdmin() bool
	SetIsLogedIn(bool)
	GetIsLogedIn() bool
	SetChanID(uint32)
	GetChanID() uint32
	AddCloseCallback(*func())
	// Below here might not be needed
	SetWorldID(uint32)
	GetWorldID() uint32
}

type Pos interface {
	SetX(int16)
	GetX() int16
	SetY(int16)
	GetY() int16
}

type Life interface {
	SetID(uint32)
	GetID() uint32
	Pos
	SetFoothold(int16)
	GetFoothold() int16
	SetFace(bool)
	GetFace() bool
	GetController() ClientConn
	SetController(ClientConn)
}

type Npc interface {
	Life
	SetRx0(int16)
	GetRx0() int16
	SetRx1(int16)
	GetRx1() int16
}

type Mob interface {
	Life
	GetEXP() uint32
	SetEXP(uint32)
	GetHp() uint16
	SetHp(uint16)
	GetMaxHp() uint16
	SetMaxHp(uint16)
	GetMp() uint16
	SetMp(uint16)
	GetMaxMp() uint16
	SetMaxMp(uint16)
	GetBoss() bool
	SetBoss(bool)
	GetLevel() byte
	SetLevel(byte)
}

type Portal interface {
	GetToMap() uint32
	SetToMap(uint32)
	GetName() string
	SetName(string)
	Pos
	GetIsSpawn() bool
	SetIsSpawn(bool)
}

type Map interface {
	GetNps() []Npc
	AddNpc(Npc)
	GetMobs() []Mob
	AddMob(Mob)
	GetPortals() []Portal
	AddPortal(Portal)
	GetPlayers() []ClientConn
	AddPlayer(ClientConn)
	RemovePlayer(ClientConn)
}

type Maps interface {
	GetMap(uint32) Map
}

type Characters interface {
	AddOnlineCharacter(OcClientConn, *character.Character)
	RemoveOnlineCharacter(OcClientConn)
	GetOnlineCharacterHandle(OcClientConn) *character.Character
	GetConnectionHandle(string) OcClientConn
}
