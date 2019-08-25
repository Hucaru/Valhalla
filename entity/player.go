package entity

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Players type alias
type Players []*Player

// GetFromConn retrieve the player from the connection
func (p *Players) GetFromConn(conn mnet.Client) (*Player, error) {
	for _, v := range *p {
		if v.conn == conn {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

// GetFromName retrieve the player from the connection
func (p *Players) GetFromName(name string) (*Player, error) {
	for _, v := range *p {
		if v.char.name == name {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

// RemoveFromConn removes the player based on the connection
func (p *Players) RemoveFromConn(conn mnet.Client) error {
	i := -1

	for j, v := range *p {
		if v.conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("Could not find player")
	}

	(*p)[i] = (*p)[len(*p)-1]
	(*p)[len(*p)-1] = nil
	(*p) = (*p)[:len(*p)-1]

	return nil
}

// Player connected to server
type Player struct {
	conn       mnet.Client
	char       Character
	instanceID int
}

func NewPlayer(conn mnet.Client, char Character) *Player {
	return &Player{conn: conn, char: char, instanceID: 0}
}

func (p Player) Char() Character {
	return p.char
}

func (p Player) InstanceID() int {
	return p.instanceID
}

func (p *Player) SetInstance(id int) {
	p.instanceID = id
}

func (p Player) Send(packet mpacket.Packet) {
	p.conn.Send(packet)
}

func (p *Player) SetJob(id int16) {
	p.char.job = id
	p.conn.Send(PacketPlayerStatChange(true, constant.JobID, int32(id)))
}

func (p *Player) levelUp(inst *instance) {
	p.GiveAP(5)
	p.GiveSP(3)

	levelUpHp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch p.char.job / 100 { // add effects from skills e.g. improve max mp
	case 0:
		p.char.maxHP += levelUpHp(constant.BeginnerHpAdd, 0)
		p.char.maxMP += levelUpMp(constant.BeginnerMpAdd, p.char.intt)
	case 1:
		p.char.maxHP += levelUpHp(constant.WarriorHpAdd, 0)
		p.char.maxMP += levelUpMp(constant.WarriorMpAdd, p.char.intt)
	case 2:
		p.char.maxHP += levelUpHp(constant.MagicianHpAdd, 0)
		p.char.maxMP += levelUpMp(constant.MagicianMpAdd, 2*p.char.intt)
	case 3:
		p.char.maxHP += levelUpHp(constant.BowmanHpAdd, 0)
		p.char.maxMP += levelUpMp(constant.BowmanMpAdd, p.char.intt)
	case 4:
		p.char.maxHP += levelUpHp(constant.ThiefHpAdd, 0)
		p.char.maxMP += levelUpMp(constant.ThiefMpAdd, p.char.intt)
	case 5:
		p.char.maxHP += constant.AdminHpAdd
		p.char.maxMP += constant.AdminMpAdd
	default:
		log.Println("Unkown job during level up", p.char.job)
	}

	p.char.hp = p.char.maxHP
	p.char.mp = p.char.maxMP

	p.SetHP(p.char.hp)
	p.SetMaxHP(p.char.hp)

	p.SetMP(p.char.mp)
	p.SetMaxMP(p.char.mp)

	p.GiveLevel(1, inst)
}

func (p *Player) SetEXP(amount int32, inst *instance) {
	if p.char.level > 199 {
		return
	}

	remainder := amount - constant.ExpTable[p.char.level-1]

	if remainder >= 0 {
		p.levelUp(inst)
		p.SetEXP(remainder, inst)
	} else {
		p.char.exp = amount
		p.Send(PacketPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

func (p *Player) GiveEXP(amount int32, fromMob, fromParty bool, inst *instance) {
	if fromMob {
		p.Send(PacketMessageExpGained(!fromParty, false, amount))
	} else {
		p.Send(PacketMessageExpGained(true, true, amount))
	}

	p.SetEXP(p.char.exp+amount, inst)
}

func (p *Player) SetLevel(amount byte, inst *instance) {
	p.char.level = amount
	p.Send(PacketPlayerStatChange(false, constant.LevelID, int32(amount)))
	inst.Send(PacketPlayerLevelUpAnimation(p.char.id))
}

func (p *Player) GiveLevel(amount byte, inst *instance) {
	p.SetLevel(p.char.level+amount, inst)
}

func (p *Player) SetAP(amount int16) {
	p.char.ap = amount
	p.Send(PacketPlayerStatChange(false, constant.ApID, int32(amount)))
}

func (p *Player) GiveAP(amount int16) {
	p.SetAP(p.char.ap + amount)
}

func (p *Player) SetSp(amount int16) {
	p.char.sp = amount
	p.Send(PacketPlayerStatChange(false, constant.SpID, int32(amount)))
}

func (p *Player) GiveSP(amount int16) {
	p.SetSp(p.char.sp + amount)
}

func (p *Player) SetStr(amount int16) {
	p.char.str = amount
	p.Send(PacketPlayerStatChange(true, constant.StrID, int32(amount)))
}

func (p *Player) GiveStr(amount int16) {
	p.SetStr(p.char.str + amount)
}

func (p *Player) SetDex(amount int16) {
	p.char.dex = amount
	p.Send(PacketPlayerStatChange(true, constant.DexID, int32(amount)))
}

func (p *Player) GiveDex(amount int16) {
	p.SetDex(p.char.dex + amount)
}

func (p *Player) SetInt(amount int16) {
	p.char.intt = amount
	p.Send(PacketPlayerStatChange(true, constant.IntID, int32(amount)))
}

func (p *Player) GiveInt(amount int16) {
	p.SetInt(p.char.intt + amount)
}

func (p *Player) SetLuk(amount int16) {
	p.char.luk = amount
	p.Send(PacketPlayerStatChange(true, constant.LukID, int32(amount)))
}

func (p *Player) GiveLuk(amount int16) {
	p.SetLuk(p.char.luk + amount)
}

func (p *Player) SetHP(amount int16) {
	p.char.hp = amount
	p.Send(PacketPlayerStatChange(true, constant.HpID, int32(amount)))
}

func (p *Player) GiveHP(amount int16) {
	p.SetHP(p.char.hp + amount)
	if p.char.hp < 0 {
		p.SetHP(0)
	}
}

func (p *Player) SetMaxHP(amount int16) {
	p.char.maxHP = amount
	p.Send(PacketPlayerStatChange(true, constant.MaxHpID, int32(amount)))
}

func (p *Player) SetMP(amount int16) {
	p.char.mp = amount
	p.Send(PacketPlayerStatChange(true, constant.HpID, int32(amount)))
}

func (p *Player) GiveMP(amount int16) {
	p.SetMP(p.char.mp + amount)
	if p.char.mp < 0 {
		p.SetMP(0)
	}
}

func (p *Player) SetMaxMP(amount int16) {
	p.char.maxMP = amount
	p.Send(PacketPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

func (p *Player) SetFame(amount int16) {

}

func (p *Player) SetGuild(name string) {

}

func (p *Player) SetEquipSlotSize(size byte) {

}

func (p *Player) SetUseSlotSize(size byte) {

}

func (p *Player) SetEtcSlotSize(size byte) {

}

func (p *Player) SetCashSlotSize(size byte) {

}

func (p *Player) SetMesos(amount int32) {
	p.char.mesos = amount
	p.Send(PacketPlayerStatChange(false, constant.MesosID, amount))
}

func (p *Player) GiveMesos(amount int32) {
	p.SetMesos(p.char.mesos + amount)
}

func (p *Player) SetMinigameWins(v int32) {
	p.char.minigameWins = v
}

func (p *Player) SetMinigameLoss(v int32) {
	p.char.minigameLoss = v
}

func (p *Player) SetMinigameDraw(v int32) {
	p.char.minigameDraw = v
}

func (p *Player) UpdateMovement(frag movementFrag) {
	p.char.pos.x = frag.x
	p.char.pos.y = frag.y
	p.char.foothold = frag.foothold
	p.char.stance = frag.stance
}

func (p *Player) SetPos(pos pos) {
	p.char.pos = pos
}

func (p Player) CheckPos(pos pos, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = p.char.pos.x == pos.x
	} else {
		xValid = (pos.x-xRange < p.char.pos.x && p.char.pos.x < pos.x+xRange)
	}

	if yRange == 0 {
		xValid = p.char.pos.y == pos.y
	} else {
		yValid = (pos.y-yRange < p.char.pos.y && p.char.pos.y < pos.y+yRange)
	}

	return xValid && yValid
}

func (p *Player) SetFoothold(fh int16) {
	p.char.foothold = fh
}

func (p *Player) SetMapID(id int32) {
	p.char.mapID = id
}

func (p *Player) SetMapPosID(pos byte) {
	p.char.mapPos = pos
}
