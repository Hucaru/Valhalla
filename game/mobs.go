package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type gameMob struct {
	def.Mob
	mapID    int32
	dmgTaken map[mnet.MConnChannel]int32
}

func createNewMob(spawnID, mapID int32, life nx.Life, info nx.Mob) gameMob {
	return gameMob{Mob: def.CreateMob(spawnID, life, info, nil),
		mapID:    mapID,
		dmgTaken: make(map[mnet.MConnChannel]int32)}
}

func (m gameMob) FacesLeft() bool {
	return m.Stance%2 != 0
}

func (m *gameMob) GiveDamage(player Player, damages []int32) {
	if m.HP > 0 && m.Controller != player.MConnChannel {
		m.ChangeController(player.MConnChannel)
	}

	for _, dmg := range damages {
		if dmg > m.HP {
			m.HP = 0
			m.dmgTaken[player.MConnChannel] += m.HP
		} else {
			m.HP -= dmg
			m.dmgTaken[player.MConnChannel] += dmg
		}
	}
}

func (m *gameMob) ChangeController(newController mnet.MConnChannel) {
	if newController == nil {
		m.Controller = nil
		return
	}

	if m.Controller == newController {
		return
	}

	if m.Controller != nil {
		m.Controller.Send(packet.MobEndControl(m.Mob))
	}

	m.Controller = newController
	newController.Send(packet.MobControl(m.Mob))
}

func (m *gameMob) FindNewControllerExcept(conn mnet.MConnChannel) {
	var newController mnet.MConnChannel

	for c, v := range Players {
		if v.char.MapID == m.mapID {
			if c == conn {
				continue
			} else {
				newController = c
			}
		}
	}

	if newController == nil {
		return
	}

	m.ChangeController(Players[newController])
}
