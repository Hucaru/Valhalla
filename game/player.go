package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type playersList map[mnet.MConnChannel]*Player

var Players = playersList{}

func (p playersList) GetFromName(name string) (*Player, error) {
	for i, v := range p {
		if v.char.Name == name {
			return p[i], nil
		}
	}

	return nil, fmt.Errorf("Unable to get player")
}

func (p playersList) GetFromConn(conn mnet.MConnChannel) (*Player, error) {
	for i := range p {
		if i == conn {
			return p[i], nil
		}
	}

	return nil, fmt.Errorf("Unable to get player")
}

func (p playersList) GetFromID(id int32) (*Player, error) {
	for i, v := range p {
		if v.char.ID == id {
			return p[i], nil
		}
	}

	return nil, fmt.Errorf("Unable to get player")
}

type Player struct {
	mnet.MConnChannel
	char                 *def.Character
	LastAttackPacketTime int64
	InstanceID           int
	RoomID               int32
}

func NewPlayer(conn mnet.MConnChannel, char def.Character) *Player {
	return &Player{MConnChannel: conn, char: &char, InstanceID: 0}
}

func (p Player) Char() def.Character {
	return *p.char
}

func (p *Player) ChangeMap(mapID int32, portal nx.Portal, portalID byte) {
	Maps[p.char.MapID].RemovePlayer(p.MConnChannel)

	p.char.Pos.X = portal.X
	p.char.Pos.Y = portal.Y
	p.char.MapPos = portalID
	p.char.MapID = mapID

	p.Send(PacketMapChange(mapID, 0, portalID, p.char.HP)) // get current channel

	Maps[p.char.MapID].AddPlayer(p.MConnChannel, p.InstanceID)
}

func (p *Player) ChangeInstance(newInstID int) {
	if newInstID >= len(Maps[p.char.MapID].instances) {
		return
	}

	Maps[p.char.MapID].RemovePlayer(p.MConnChannel)

	p.InstanceID = newInstID

	portal, portalID, _ := Maps[p.char.MapID].GetRandomSpawnPortal()
	p.char.Pos.X = portal.X
	p.char.Pos.Y = portal.Y
	p.char.MapPos = portalID

	p.Send(PacketMapChange(p.char.MapID, 0, portalID, p.char.HP)) // get current channel

	Maps[p.char.MapID].AddPlayer(p.MConnChannel, p.InstanceID)
}

func (p *Player) UpdateMovement(moveData def.MovementFrag) {
	p.char.Pos.X = moveData.X
	p.char.Pos.Y = moveData.Y
	// p.char.Foothold = moveData.Foothold - makes char warp accross map to other players when going through portal
	p.char.Stance = moveData.Stance
}

func (p *Player) Kill() {
	p.SetHP(0)
}

func (p *Player) Revive() {
	p.SetHP(int32(p.char.MaxHP))
}

func (p *Player) SetMaxHP(amount int32) {
	if amount > math.MaxInt16 {
		amount = math.MaxInt16
	}

	p.char.MaxHP = int16(amount)
	p.Send(PacketPlayerStatChange(true, constant.MaxHpID, amount))
}

func (p *Player) SetHP(amount int32) {
	p.char.HP = int16(amount)

	if p.char.HP > p.char.MaxHP {
		p.char.HP = p.char.MaxHP
	}

	if p.char.HP < 0 {
		p.char.HP = 0
	}

	p.Send(PacketPlayerStatChange(true, constant.HpID, amount))
}

func (p *Player) GiveHP(amount int32) {
	p.SetHP(int32(p.char.HP) + amount)
	if p.char.HP < 1 {
		p.Kill()
	}
}

func (p *Player) SetMaxMP(amount int32) {
	if amount > math.MaxInt16 {
		amount = math.MaxInt16
	}

	p.char.MaxMP = int16(amount)
	p.Send(PacketPlayerStatChange(true, constant.MaxMpID, amount))
}

func (p *Player) SetMP(amount int32) {
	p.char.MP = int16(amount)

	if p.char.MP > p.char.MaxMP {
		p.char.MP = p.char.MaxMP
	}

	if p.char.MP < 0 {
		p.char.MP = 0
	}

	p.Send(PacketPlayerStatChange(true, constant.MpID, amount))
}

func (p *Player) GiveMP(amount int32) {
	p.SetMP(int32(p.char.MP) + amount)
}

func (p *Player) SetJob(jobID int16) {
	p.char.Job = jobID
	p.Send(PacketPlayerStatChange(true, constant.JobID, int32(jobID)))
}

func (p *Player) levelUp() {
	p.GiveAP(5)
	p.GiveSP(3)

	levelUpHp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch p.char.Job / 100 { // add effects from skills e.g. improve max mp
	case 0:
		p.char.MaxHP += levelUpHp(constant.BeginnerHpAdd, 0)
		p.char.MaxMP += levelUpMp(constant.BeginnerMpAdd, p.char.Int)
	case 1:
		p.char.MaxHP += levelUpHp(constant.WarriorHpAdd, 0)
		p.char.MaxMP += levelUpMp(constant.WarriorMpAdd, p.char.Int)
	case 2:
		p.char.MaxHP += levelUpHp(constant.MagicianHpAdd, 0)
		p.char.MaxMP += levelUpMp(constant.MagicianMpAdd, 2*p.char.Int)
	case 3:
		p.char.MaxHP += levelUpHp(constant.BowmanHpAdd, 0)
		p.char.MaxMP += levelUpMp(constant.BowmanMpAdd, p.char.Int)
	case 4:
		p.char.MaxHP += levelUpHp(constant.ThiefHpAdd, 0)
		p.char.MaxMP += levelUpMp(constant.ThiefMpAdd, p.char.Int)
	case 5:
		p.char.MaxHP += constant.AdminHpAdd
		p.char.MaxMP += constant.AdminMpAdd
	default:
		fmt.Println("Unkown job", p.char.Job)
	}

	p.char.HP = p.char.MaxHP
	p.char.MP = p.char.MaxMP

	p.SetHP(int32(p.char.HP))
	p.SetMaxHP(int32(p.char.HP))

	p.SetMP(int32(p.char.MP))
	p.SetMaxMP(int32(p.char.MP))

	p.GiveLevel(1)
}

func (p *Player) SetEXP(amount int32) {
	if p.char.Level > 199 {
		return
	}

	remainder := amount - constant.ExpTable[p.char.Level-1]
	if remainder >= 0 {
		p.levelUp()
		p.SetEXP(remainder)
	} else {
		p.char.EXP = amount
		p.Send(PacketPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

func (p *Player) GiveEXP(amount int32, fromMob, fromParty bool) {
	if fromMob {
		p.Send(PacketMessageExpGained(!fromParty, false, amount))
	} else {
		p.Send(PacketMessageExpGained(true, true, amount))
	}

	p.SetEXP(p.char.EXP + amount)
}

func (p *Player) SetLevel(level byte) {
	p.char.Level = level
	p.Send(PacketPlayerStatChange(false, constant.LevelID, int32(level)))
	Maps[p.char.MapID].Send(PacketPlayerLevelUpAnimation(p.char.ID), p.InstanceID)
}

func (p *Player) GiveLevel(amount int8) {
	p.SetLevel(byte(int8(p.char.Level) + amount))
}

func (p *Player) SetAP(amount int16) {
	p.char.AP = amount
	p.Send(PacketPlayerStatChange(false, constant.ApID, int32(amount)))
}

func (p *Player) GiveAP(amount int16) {
	p.SetAP(p.char.AP + amount)
}

func (p *Player) SetSP(amount int16) {
	p.char.SP = amount
	p.Send(PacketPlayerStatChange(false, constant.SpID, int32(amount)))
}

func (p *Player) GiveSP(amount int16) {
	p.SetSP(p.char.SP + amount)
}

func (p *Player) SetStr(amount int16) {
	p.char.Str = amount
	p.Send(PacketPlayerStatChange(true, constant.StrID, int32(amount)))
}

func (p *Player) GiveStr(amount int16) {
	p.SetStr(p.char.Str + amount)
}

func (p *Player) SetDex(amount int16) {
	p.char.Dex = amount
	p.Send(PacketPlayerStatChange(true, constant.DexID, int32(amount)))
}

func (p *Player) GiveDex(amount int16) {
	p.SetDex(p.char.Dex + amount)
}

func (p *Player) SetInt(amount int16) {
	p.char.Int = amount
	p.Send(PacketPlayerStatChange(true, constant.IntID, int32(amount)))
}

func (p *Player) GiveInt(amount int16) {
	p.SetInt(p.char.Int + amount)
}

func (p *Player) SetLuk(amount int16) {
	p.char.Luk = amount
	p.Send(PacketPlayerStatChange(true, constant.LukID, int32(amount)))
}

func (p *Player) GiveLuk(amount int16) {
	p.SetLuk(p.char.Luk + amount)
}

func (p *Player) SetMesos(amount int32) {
	p.char.Mesos = amount
	p.Send(PacketPlayerStatChange(false, constant.MesosID, amount))
}

func (p *Player) GiveMesos(amount int32) {
	p.SetMesos(p.char.Mesos + amount)
}

func (p *Player) SetMinigameWins(v int32) {
	p.char.MiniGameWins = v
}

func (p *Player) SetMinigameLoss(v int32) {
	p.char.MiniGameLoss = v
}

func (p *Player) SetMinigameDraw(v int32) {
	p.char.MiniGameDraw = v
}

func (p *Player) UpdateSkill(skill def.Skill) {
	p.char.Skills[skill.ID] = skill
	p.Send(PacketPlayerSkillBookUpdate(skill.ID, int32(skill.Level)))
}
