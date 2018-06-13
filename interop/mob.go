package interop

type Mob interface {
	Npc
	GetFlySpeed() uint32
	GetSummoner() ClientConn
}
