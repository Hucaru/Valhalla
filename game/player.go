package game

import (
	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

var players = map[mnet.MConnChannel]Player{}

type Player struct {
	mnet.MConnChannel
	char                 *types.Character
	LastAttackPacketTime int64
}

func NewPlayer(conn mnet.MConnChannel, char types.Character) Player {
	return Player{MConnChannel: conn, char: &char}
}

func (p Player) Char() types.Character {
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

func (p *Player) UpdateMovement(moveData types.MovementFrag) {
	p.char.Pos.X = moveData.X
	p.char.Pos.Y = moveData.Y
	// p.char.Foothold = moveData.Foothold - makes char warp accross map to other players when going through portal
	p.char.Stance = moveData.Stance
}

func (p *Player) Kill() {

}

func (p *Player) SetHP(ammount int16) {
	p.char.HP = ammount

	if p.char.HP > p.char.MaxHP {
		p.char.HP = p.char.MaxHP
	}

	p.Send(packets.PlayerStatChange(true, consts.HP_ID, int32(p.char.HP)))
}

func (p *Player) GiveHP(ammount int16) {
	p.SetHP(p.char.HP + ammount)
}

func (p *Player) SetMP(ammount int16) {
	p.char.MP = ammount

	if p.char.MP > p.char.MaxMP {
		p.char.MP = p.char.MaxMP
	}

	p.Send(packets.PlayerStatChange(true, consts.MP_ID, int32(p.char.MP)))
}

func (p *Player) GiveMP(ammount int16) {
	p.SetMP(p.char.MP + ammount)
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

func (p *Player) SetMesos() {

}

func (p *Player) GiveMesos() {

}

func (p *Player) GiveItem() {

}

func (p *Player) TakeItem() {

}
