package entity

import "github.com/Hucaru/Valhalla/mpacket"

type Pos struct {
	X int16
	Y int16
}

type AttackInfo struct {
	SpawnID                                                int32
	HitAction, ForeAction, FrameIndex, CalcDamageStatIndex byte
	FacesLeft                                              bool
	HitPosition, PreviousMobPosition                       Pos
	HitDelay                                               int16
	Damages                                                []int32
}

type AttackData struct {
	SkillID, SummonType, TotalDamage, ProjectileID int32
	IsMesoExplosion, FacesLeft                     bool
	Option, Action, AttackType                     byte
	Targets, Hits, SkillLevel                      byte

	AttackInfo []AttackInfo
	PlayerPos  Pos
}

func (data *AttackData) Parse(reader mpacket.Reader) {

}
