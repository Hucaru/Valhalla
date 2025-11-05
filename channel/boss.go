package channel

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/Hucaru/Valhalla/constant"
)

func manageSummonedBoss(inst *fieldInstance, mobID int32, server *Server) {
	log.Println("Boss handler spawned for", mobID)

	inst.dispatch <- func() {
		inst.properties["eventActive"] = true
	}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	timeout := time.NewTimer(time.Hour * 1)
	defer timeout.Stop()

	var finished atomic.Bool
	finished.Store(false)

	for {
		select {
		case <-ticker.C:
			if finished.Load() {
				return
			}

			inst.dispatch <- func() {
				for _, mob := range inst.lifePool.mobs {
					if mob.boss {
						return
					}
				}

				inst.properties["eventActive"] = false
				finished.Store(true)
			}
		case <-timeout.C:
			var returnMap int32

			switch inst.fieldID {
			case constant.MapBossPapulatus:
				returnMap = constant.MapBossPapulatusReturn
			case constant.MapBossZakum:
				returnMap = constant.MapBossZakumReturn
			}

			inst.dispatch <- func() {
				field, ok := server.fields[returnMap]

				if !ok {
					log.Println("Error in getting field")
					return
				}

				dest, err := field.getInstance(0)

				if err != nil {
					log.Println(err)
					return
				}

				portal, err := dest.getPortalFromID(0)

				if err != nil {
					log.Println(err)
					return
				}

				for _, plr := range inst.players {
					server.warpPlayer(plr, field, portal, true)
				}

				inst.lifePool.eraseMobs()
				inst.reactorPool.reset(false)
				inst.properties["eventActive"] = false
			}
			return
		}
	}
}

func summonRequiresBossHandler(mobID int32) bool {
	switch mobID {
	case constant.MobPapalatusBall:
		fallthrough
	case constant.MobZakum1Body:
		return true
	default:
		return false
	}
}
