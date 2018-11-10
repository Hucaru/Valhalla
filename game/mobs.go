package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

type mob struct {
	types.Mob
	mapID int32
}

func (m *mob) GiveDamage(player Player, damages []int32) {
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

func (m *mob) ChangeController(newController Player) {
	if m.Controller == newController {
		return
	}

	if m.Controller != nil {
		m.Controller.Send(packets.MobEndControl(m.Mob))
	}

	m.Controller = newController.MConnChannel
	newController.Send(packets.MobControl(m.Mob))
}

func (m *mob) findNewControllerExcept(mapID int32, conn mnet.MConnChannel) mnet.MConnChannel {
	return nil
}

func (m *mob) FindNewControllerExcept(conn mnet.MConnChannel) {
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

	conn.Send(packets.MobEndControl(m.Mob))
	m.Controller = newController
	m.Controller.Send(packets.MobControl(m.Mob))
}
