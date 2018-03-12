package data

type mapleMob struct {
	mapleNpc
	exp                  uint32
	hp, maxHp, mp, maxMp uint16
	boss                 bool
	level, state         byte
	mobTime              uint32
	x, y                 int16
}

func (m *mapleMob) GetEXP() uint32            { return m.exp }
func (m *mapleMob) SetEXP(exp uint32)         { m.exp = exp }
func (m *mapleMob) GetHp() uint16             { return m.hp }
func (m *mapleMob) SetHp(hp uint16)           { m.hp = hp }
func (m *mapleMob) GetMaxHp() uint16          { return m.maxHp }
func (m *mapleMob) SetMaxHp(maxHp uint16)     { m.maxHp = maxHp }
func (m *mapleMob) GetMp() uint16             { return m.mp }
func (m *mapleMob) SetMp(mp uint16)           { m.mp = mp }
func (m *mapleMob) GetMaxMp() uint16          { return m.maxMp }
func (m *mapleMob) SetMaxMp(maxMp uint16)     { m.maxMp = maxMp }
func (m *mapleMob) GetBoss() bool             { return m.boss }
func (m *mapleMob) SetBoss(boss bool)         { m.boss = boss }
func (m *mapleMob) GetLevel() byte            { return m.level }
func (m *mapleMob) SetLevel(level byte)       { m.level = level }
func (m *mapleMob) GetState() byte            { return m.state }
func (m *mapleMob) SetState(state byte)       { m.state = state }
func (m *mapleMob) GetMobTime() uint32        { return m.mobTime }
func (m *mapleMob) SetMobTime(mobTime uint32) { m.mobTime = mobTime }
