package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
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

	return &Player{}, fmt.Errorf("Unable to get player")
}

func (p playersList) GetFromConn(conn mnet.MConnChannel) (*Player, error) {
	for i := range p {
		if i == conn {
			return p[i], nil
		}
	}

	return &Player{}, fmt.Errorf("Unable to get player")
}

func (p playersList) GetFromID(id int32) (*Player, error) {
	for i, v := range p {
		if v.char.ID == id {
			return p[i], nil
		}
	}

	return &Player{}, fmt.Errorf("Unable to get player")
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

	p.Send(packet.MapChange(mapID, 0, portalID, p.char.HP)) // get current channel

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

	p.Send(packet.MapChange(p.char.MapID, 0, portalID, p.char.HP)) // get current channel

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

func (p *Player) SetMaxHP(ammount int32) {
	if ammount > math.MaxInt16 {
		ammount = math.MaxInt16
	}

	p.char.MaxHP = int16(ammount)
	p.Send(packet.PlayerStatChange(true, constant.MaxHpId, ammount))
}

func (p *Player) SetHP(ammount int32) {
	p.char.HP = int16(ammount)

	if p.char.HP > p.char.MaxHP {
		p.char.HP = p.char.MaxHP
	}

	if p.char.HP < 0 {
		p.char.HP = 0
	}

	p.Send(packet.PlayerStatChange(true, constant.HpId, ammount))
}

func (p *Player) GiveHP(ammount int32) {
	p.SetHP(int32(p.char.HP) + ammount)
}

func (p *Player) SetMaxMP(ammount int32) {
	if ammount > math.MaxInt16 {
		ammount = math.MaxInt16
	}

	p.char.MaxMP = int16(ammount)
	p.Send(packet.PlayerStatChange(true, constant.MaxMpId, ammount))
}

func (p *Player) SetMP(ammount int32) {
	p.char.MP = int16(ammount)

	if p.char.MP > p.char.MaxMP {
		p.char.MP = p.char.MaxMP
	}

	if p.char.MP < 0 {
		p.char.MP = 0
	}

	p.Send(packet.PlayerStatChange(true, constant.MpId, ammount))
}

func (p *Player) GiveMP(ammount int32) {
	p.SetMP(int32(p.char.MP) + ammount)
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

func (p *Player) SetEXP(ammount int32) {
	remainder := ammount - constant.ExpTable[p.char.Level-1]
	if remainder >= 0 {
		p.levelUp()
		p.SetEXP(remainder)
	} else {
		p.char.EXP = ammount
		p.Send(packet.PlayerStatChange(false, constant.ExpId, int32(ammount)))
	}
}

func (p *Player) GiveEXP(ammount int32) {
	p.SetEXP(p.char.EXP + ammount)
}

func (p *Player) SetLevel(level byte) {
	p.char.Level += 1
	p.Send(packet.PlayerStatChange(false, constant.LevelId, int32(level)))
	Maps[p.char.MapID].Send(packet.PlayerLevelUpAnimation(p.char.ID), p.InstanceID)
}

func (p *Player) GiveLevel(ammount byte) {
	p.SetLevel(p.char.Level + ammount)
}

func (p *Player) SetAP(ammount int16) {
	p.char.AP = ammount
	p.Send(packet.PlayerStatChange(false, constant.ApId, int32(ammount)))
}

func (p *Player) GiveAP(ammount int16) {
	p.SetAP(p.char.AP + ammount)
}

func (p *Player) SetSP(ammount int16) {
	p.char.SP = ammount
	p.Send(packet.PlayerStatChange(false, constant.SpId, int32(ammount)))
}

func (p *Player) GiveSP(ammount int16) {
	p.SetSP(p.char.SP + ammount)
}

func (p *Player) SetStr(ammount int16) {
	p.char.Str = ammount
	p.Send(packet.PlayerStatChange(true, constant.StrId, int32(ammount)))
}

func (p *Player) GiveStr(ammount int16) {
	p.SetStr(p.char.Str + ammount)
}

func (p *Player) SetDex(ammount int16) {
	p.char.Dex = ammount
	p.Send(packet.PlayerStatChange(true, constant.DexId, int32(ammount)))
}

func (p *Player) GiveDex(ammount int16) {
	p.SetDex(p.char.Dex + ammount)
}

func (p *Player) SetInt(ammount int16) {
	p.char.Int = ammount
	p.Send(packet.PlayerStatChange(true, constant.IntId, int32(ammount)))
}

func (p *Player) GiveInt(ammount int16) {
	p.SetInt(p.char.Int + ammount)
}

func (p *Player) SetLuk(ammount int16) {
	p.char.Luk = ammount
	p.Send(packet.PlayerStatChange(true, constant.LukId, int32(ammount)))
}

func (p *Player) GiveLuk(ammount int16) {
	p.SetLuk(p.char.Luk + ammount)
}

func (p *Player) SetMesos(ammount int32) {
	p.char.Mesos = ammount
	p.Send(packet.PlayerStatChange(false, constant.MesosId, ammount))
}

func (p *Player) GiveMesos(ammount int32) {
	p.SetMesos(p.char.Mesos + ammount)
}

func (p *Player) GiveItem() {

}

func (p *Player) TakeItem() {

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
