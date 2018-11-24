package game

import (
	"math"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/def"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
)

var players = map[mnet.MConnChannel]Player{}

type Player struct {
	mnet.MConnChannel
	char                 *def.Character
	LastAttackPacketTime int64
}

func NewPlayer(conn mnet.MConnChannel, char def.Character) Player {
	return Player{MConnChannel: conn, char: &char}
}

func (p Player) Char() def.Character {
	return *p.char
}

func (p *Player) ChangeMap(mapID int32, portal nx.Portal, portalID byte) {
	for _, player := range players {
		if player.Char().CurrentMap == p.char.CurrentMap {
			player.Send(packets.MapPlayerLeft(p.char.ID))
		}
	}

	maps[p.char.CurrentMap].removeController(p.MConnChannel)

	p.char.Pos.X = portal.X
	p.char.Pos.Y = portal.Y
	p.char.CurrentMapPos = portalID
	p.char.CurrentMap = mapID

	p.Send(packets.MapChange(mapID, 0, portalID, p.char.HP)) // get current channel
	p.sendMapItems()

	maps[p.char.CurrentMap].addController(p.MConnChannel)
}

func (p *Player) sendMapItems() {
	for _, mob := range maps[p.char.CurrentMap].mobs {
		if mob.HP > 0 {
			mob.SummonType = -1 // -2: fade in spawn animation, -1: no spawn animation
			p.Send(packets.MobShow(mob.Mob))
		}
	}

	for _, npc := range maps[p.char.CurrentMap].npcs {
		p.Send(packets.NpcShow(npc))
		p.Send(packets.NPCSetController(npc.SpawnID, true))
	}

	for _, player := range players {
		if player.Char().CurrentMap == p.char.CurrentMap {
			player.Send(packets.MapPlayerEnter(p.Char()))
			p.Send(packets.MapPlayerEnter(player.Char()))
		}
	}
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
	p.Send(packets.PlayerStatChange(true, consts.MAX_HP_ID, ammount))
}

func (p *Player) SetHP(ammount int32) {
	p.char.HP = int16(ammount)

	if p.char.HP > p.char.MaxHP {
		p.char.HP = p.char.MaxHP
	}

	if p.char.HP < 0 {
		p.char.HP = 0
	}

	p.Send(packets.PlayerStatChange(true, consts.HP_ID, ammount))
}

func (p *Player) GiveHP(ammount int32) {
	p.SetHP(int32(p.char.HP) + ammount)
}

func (p *Player) SetMaxMP(ammount int32) {
	if ammount > math.MaxInt16 {
		ammount = math.MaxInt16
	}

	p.char.MaxMP = int16(ammount)
	p.Send(packets.PlayerStatChange(true, consts.MAX_MP_ID, ammount))
}

func (p *Player) SetMP(ammount int32) {
	p.char.MP = int16(ammount)

	if p.char.MP > p.char.MaxMP {
		p.char.MP = p.char.MaxMP
	}

	if p.char.MP < 0 {
		p.char.MP = 0
	}

	p.Send(packets.PlayerStatChange(true, consts.MP_ID, ammount))
}

func (p *Player) GiveMP(ammount int32) {
	p.SetMP(int32(p.char.MP) + ammount)
}

func (p *Player) SetEXP() {

}

func (p *Player) GiveEXP() {

}

func (p *Player) SetLevel() {

}

func (p *Player) GiveLevel() {
}

func (p *Player) SetAP(ammount int16) {

}

func (p *Player) GiveAP(ammount int16) {
	p.SetAP(p.char.AP + ammount)
}

func (p *Player) SetSP(ammount int16) {

}

func (p *Player) GiveSP(ammount int16) {
	p.SetSP(p.char.SP + ammount)
}

func (p *Player) SetStr(ammount int16) {

}

func (p *Player) GiveStr(ammount int16) {
	p.SetStr(p.char.Str + ammount)
}

func (p *Player) SetDex(ammount int16) {

}

func (p *Player) GiveDex(ammount int16) {
	p.SetDex(p.char.Dex + ammount)
}

func (p *Player) SetInt(ammount int16) {

}

func (p *Player) GiveInt(ammount int16) {
	p.SetInt(p.char.Int + ammount)
}

func (p *Player) SetLuk(ammount int16) {

}

func (p *Player) GiveLuk(ammount int16) {
	p.SetLuk(p.char.Luk + ammount)
}

func (p *Player) SetMesos() {

}

func (p *Player) GiveMesos() {

}

func (p *Player) GiveItem() {

}

func (p *Player) TakeItem() {

}
