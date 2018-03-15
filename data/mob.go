package data

import (
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/nx"
)

type mapleMob struct {
	mapleNpc
	exp                  uint32
	hp, maxHp, mp, maxMp uint32
	boss, respawns       bool
	level, state         byte
	x, y                 int16
	mobTime, deathTime   int64
	dmgReceived          map[interfaces.ClientConn]uint32
}

func (m *mapleMob) GetEXP() uint32                                       { return m.exp }
func (m *mapleMob) SetEXP(exp uint32)                                    { m.exp = exp }
func (m *mapleMob) GetHp() uint32                                        { return m.hp }
func (m *mapleMob) SetHp(hp uint32)                                      { m.hp = hp }
func (m *mapleMob) GetMaxHp() uint32                                     { return m.maxHp }
func (m *mapleMob) SetMaxHp(maxHp uint32)                                { m.maxHp = maxHp }
func (m *mapleMob) GetMp() uint32                                        { return m.mp }
func (m *mapleMob) SetMp(mp uint32)                                      { m.mp = mp }
func (m *mapleMob) GetMaxMp() uint32                                     { return m.maxMp }
func (m *mapleMob) SetMaxMp(maxMp uint32)                                { m.maxMp = maxMp }
func (m *mapleMob) GetBoss() bool                                        { return m.boss }
func (m *mapleMob) SetBoss(boss bool)                                    { m.boss = boss }
func (m *mapleMob) GetLevel() byte                                       { return m.level }
func (m *mapleMob) SetLevel(level byte)                                  { m.level = level }
func (m *mapleMob) GetState() byte                                       { return m.state }
func (m *mapleMob) SetState(state byte)                                  { m.state = state }
func (m *mapleMob) GetMobTime() int64                                    { return m.mobTime }
func (m *mapleMob) SetMobTime(mobTime int64)                             { m.mobTime = mobTime }
func (m *mapleMob) SetDeathTime(mobTime int64)                           { m.deathTime = mobTime }
func (m *mapleMob) GetDeathTime() int64                                  { return m.deathTime }
func (m *mapleMob) GetRespawns() bool                                    { return m.respawns }
func (m *mapleMob) SetRespawns(respawns bool)                            { m.respawns = respawns }
func (m *mapleMob) SetDmgReceived(dmgR map[interfaces.ClientConn]uint32) { m.dmgReceived = dmgR }
func (m *mapleMob) GetDmgReceived() map[interfaces.ClientConn]uint32     { return m.dmgReceived }

func CreateMobFromID(mobID uint32) interfaces.Mob {
	l := mapleMob{}

	mon := nx.Mob[mobID]

	l.SetID(mobID)
	l.SetBoss(mon.Boss)
	l.SetMobTime(0)
	l.SetEXP(mon.Exp)
	l.SetMaxHp(mon.MaxHp)
	l.SetHp(mon.MaxHp)
	l.SetMaxMp(mon.MaxMp)
	l.SetMp(mon.MaxMp)
	l.SetLevel(mon.Level)

	return &l
}
