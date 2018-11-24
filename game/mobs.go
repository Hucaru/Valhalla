package game

import (
	"github.com/Hucaru/Valhalla/def"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
)

type gameMob struct {
	def.Mob
	mapID int32
}

func (m gameMob) FacesLeft() bool {
	return m.Stance%2 != 0
}

func (m *gameMob) GiveDamage(player Player, damages []int32) {
	if m.HP > 0 && m.Controller != player {
		m.ChangeController(player)
	}

	for _, dmg := range damages {
		if dmg > m.HP {
			m.HP = 0
		} else {
			m.HP -= dmg
		}
	}
}

func (m *gameMob) ChangeController(newController Player) {
	if m.Controller == newController {
		return
	}

	if m.Controller != nil {
		m.Controller.Send(packets.MobEndControl(m.Mob))
	}

	m.Controller = newController.MConnChannel
	newController.Send(packets.MobControl(m.Mob))
}

func (m *gameMob) FindNewControllerExcept(conn mnet.MConnChannel) {
	var newController mnet.MConnChannel

	for c, v := range players {
		if v.char.CurrentMap == m.mapID {
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

	m.ChangeController(players[newController])
}
