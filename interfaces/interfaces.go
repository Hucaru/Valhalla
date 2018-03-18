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
	AddCloseCallback(func())
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

type FragObj interface {
	Pos
	GetState() byte
	SetState(byte)
	SetFoothold(int16)
	GetFoothold() int16
}

type Life interface {
	SetID(uint32)
	GetID() uint32
	SetSpawnID(uint32)
	GetSpawnID() uint32
	Pos

	SetSX(int16)
	GetSX() int16
	SetSY(int16)
	GetSY() int16

	SetFoothold(int16)
	GetFoothold() int16
	SetSFoothold(int16)
	GetSFoothold() int16
	SetFace(byte)
	GetFace() byte
	GetState() byte
	SetState(byte)
	GetController() ClientConn
	SetController(ClientConn)

	SetIsAlive(alive bool)
	GetIsAlive() bool
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
	GetHp() uint32
	SetHp(uint32)
	GetMaxHp() uint32
	SetMaxHp(uint32)
	GetMp() uint32
	SetMp(uint32)
	GetMaxMp() uint32
	SetMaxMp(uint32)
	GetBoss() bool
	SetBoss(bool)
	GetLevel() byte
	SetLevel(byte)
	GetMobTime() int64
	SetMobTime(int64)
	SetDeathTime(int64)
	GetDeathTime() int64
	GetRespawns() bool
	SetRespawns(bool)
	SetDmgReceived(map[ClientConn]uint32)
	GetDmgReceived() map[ClientConn]uint32
}

type Portal interface {
	GetToMap() uint32
	SetToMap(uint32)
	GetToPortal() string
	SetToPortal(string)
	GetName() string
	SetName(string)
	Pos
	GetIsSpawn() bool
	SetIsSpawn(bool)
}

type Map interface {
	GetNpcs() []Npc
	AddNpc(Npc)
	GetMobs() []Mob
	GetNextMobSpawnID() uint32
	AddMob(Mob)
	RemoveMob(Mob)
	GetMobFromID(uint32) Mob
	GetReturnMap() uint32
	SetReturnMap(uint32)
	GetPortals() []Portal
	AddPortal(Portal)
	GetPlayers() []ClientConn
	AddPlayer(ClientConn)
	RemovePlayer(ClientConn)
	GetNumberSpawnableMobs() int
	GetRandomSpawnableMob(int16, int16, int16) Mob
	GetMobRate() float64
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
