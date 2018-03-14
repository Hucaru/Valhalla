package data

import (
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/nx"
)

type mapleMob struct {
	mapleNpc
	exp                  uint32
	hp, maxHp, mp, maxMp uint16
	boss                 bool
	level, state         byte
	x, y                 int16
	mobTime, deathTime   int64
}

func (m *mapleMob) GetEXP() uint32             { return m.exp }
func (m *mapleMob) SetEXP(exp uint32)          { m.exp = exp }
func (m *mapleMob) GetHp() uint16              { return m.hp }
func (m *mapleMob) SetHp(hp uint16)            { m.hp = hp }
func (m *mapleMob) GetMaxHp() uint16           { return m.maxHp }
func (m *mapleMob) SetMaxHp(maxHp uint16)      { m.maxHp = maxHp }
func (m *mapleMob) GetMp() uint16              { return m.mp }
func (m *mapleMob) SetMp(mp uint16)            { m.mp = mp }
func (m *mapleMob) GetMaxMp() uint16           { return m.maxMp }
func (m *mapleMob) SetMaxMp(maxMp uint16)      { m.maxMp = maxMp }
func (m *mapleMob) GetBoss() bool              { return m.boss }
func (m *mapleMob) SetBoss(boss bool)          { m.boss = boss }
func (m *mapleMob) GetLevel() byte             { return m.level }
func (m *mapleMob) SetLevel(level byte)        { m.level = level }
func (m *mapleMob) GetState() byte             { return m.state }
func (m *mapleMob) SetState(state byte)        { m.state = state }
func (m *mapleMob) GetMobTime() int64          { return m.mobTime }
func (m *mapleMob) SetMobTime(mobTime int64)   { m.mobTime = mobTime }
func (m *mapleMob) SetDeathTime(mobTime int64) { m.deathTime = mobTime }
func (m *mapleMob) GetDeathTime() int64        { return m.deathTime }

func CreateMobFromID(mobID uint32) interfaces.Mob {
	l := mapleMob{}

	mon := nx.Mob[mobID]

	l.SetID(mobID)
	l.SetBoss(false) // don't want it to respawn on maps that happen to contain the same mob
	l.SetMobTime(0)  // don't want it to respawn on maps that happen to contain the same mob
	l.SetEXP(mon.Exp)
	l.SetMaxHp(mon.MaxHp)
	l.SetHp(mon.MaxHp)
	l.SetMaxMp(mon.MaxMp)
	l.SetMp(mon.MaxMp)
	l.SetLevel(mon.Level)

	return &l
}
