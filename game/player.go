package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

var players = map[mnet.MConnChannel]Player{}

type Player struct {
	mnet.MConnChannel
	char *types.Character
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

	p.char.Pos.X = portal.X
	p.char.Pos.Y = portal.Y
	p.char.CurrentMapPos = portalID
	p.char.CurrentMap = mapID

	p.Send(packets.MapChange(mapID, 0, portalID, p.char.HP)) // get current channel

	p.sendMapItems()
}

func (p *Player) sendMapItems() {
	for _, npc := range maps[p.char.CurrentMap].npcs {
		p.Send(packets.NpcShow(npc))
	}

	for _, mob := range maps[p.char.CurrentMap].mobs {
		p.Send(packets.MobShow(mob, false))
	}

	for _, player := range players {
		if player.Char().CurrentMap == p.char.CurrentMap {
			player.Send(packets.MapPlayerEnter(p.Char()))
			p.Send(packets.MapPlayerEnter(player.Char()))
		}
	}
}
