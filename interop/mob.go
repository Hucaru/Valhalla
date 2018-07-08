package interop

type Mob interface {
	Npc
	GetFlySpeed() int32
	GetSummoner() ClientConn
}
