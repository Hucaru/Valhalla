package types

type AttackInfo struct {
	SpawnID                                                int32
	HitAction, ForeAction, FrameIndex, CalcDamageStatIndex byte
	FacesLeft                                              bool
	HitPosition, PreviousMobPosition                       Pos
	HitDelay                                               int16
	Damages                                                []int32
}

type AttackData struct {
	SkillID, SummonType, TotalDamage, StarID int32
	IsMesoExplosion, FacesLeft               bool
	Option, Action, AttackType               byte
	Targets, Hits, SkillLevel                byte

	AttackInfo []AttackInfo
	PlayerPos  Pos
}
