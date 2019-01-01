package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
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

func (m *gameMob) ChangeController(newController mnet.MConnChannel) {
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
