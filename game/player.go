package game

import (
	"fmt"
	"math"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

var Players = map[mnet.MConnChannel]*Player{}

type Player struct {
	mnet.MConnChannel
	char                 *def.Character
	LastAttackPacketTime int64
	InstanceID           int
}

func GetPlayerFromName(name string) (*Player, error) {
	for i, v := range Players {
		if v.char.Name == name {
			return Players[i], nil
		}
	}

	return &Player{}, fmt.Errorf("Unable to get player")
}

func GetPlayerFromID(id int32) (*Player, error) {
	for i, v := range Players {
		if v.char.ID == id {
			return Players[i], nil
		}
	}

	return &Player{}, fmt.Errorf("Unable to get player")
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

	portal, portalID := Maps[p.char.MapID].GetRandomSpawnPortal()
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
	p.Send(packet.PlayerStatChange(true, constant.MAX_HP_ID, ammount))
}

func (p *Player) SetHP(ammount int32) {
	p.char.HP = int16(ammount)

	if p.char.HP > p.char.MaxHP {
		p.char.HP = p.char.MaxHP
	}

	if p.char.HP < 0 {
		p.char.HP = 0
	}

	p.Send(packet.PlayerStatChange(true, constant.HP_ID, ammount))
}

func (p *Player) GiveHP(ammount int32) {
	p.SetHP(int32(p.char.HP) + ammount)
}

func (p *Player) SetMaxMP(ammount int32) {
	if ammount > math.MaxInt16 {
		ammount = math.MaxInt16
	}

	p.char.MaxMP = int16(ammount)
	p.Send(packet.PlayerStatChange(true, constant.MAX_MP_ID, ammount))
}

func (p *Player) SetMP(ammount int32) {
	p.char.MP = int16(ammount)

	if p.char.MP > p.char.MaxMP {
		p.char.MP = p.char.MaxMP
	}

	if p.char.MP < 0 {
		p.char.MP = 0
	}

	p.Send(packet.PlayerStatChange(true, constant.MP_ID, ammount))
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
