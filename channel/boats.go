package channel

import (
	"log"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/constant"
)

func scheduleBoats(server *Server) {
	roundUpTime := func(t time.Time, roundOn time.Duration) time.Time {
		t = t.Round(roundOn)

		if time.Since(t) >= 0 {
			t = t.Add(roundOn)
		}

		return t
	}

	wait := func(duration time.Duration) {
		timer := time.NewTimer(duration)
		<-timer.C
		timer.Stop()
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	server.dispatch <- func() {
		boatsBoarding(server, true)
	}

	// Boats take off and land every 15 mins from the hour, with boarding starting 5 mins before
	for {
		now := time.Now()
		delay := roundUpTime(now, 15*time.Minute).Sub(now)

		log.Println("Boat takeoff and landing in", delay)
		wait(delay) // time until takeoff and landing (aligned to every 15 mins on the hour)

		server.dispatch <- func() {
			arrivals := map[int32]int32{
				constant.MapBoatElliniaFlight:           constant.MapStationOrbis,
				constant.MapBoatElliniaFlightCabin:      constant.MapStationOrbis,
				constant.MapBoatOrbisElliniaFlight:      constant.MapStationEllinia,
				constant.MapBoatOrbisElliniaFlightCabin: constant.MapStationEllinia,
				constant.MapBoatOrbisLudiFlight:         constant.MapStationLudi,
				constant.MapBoatLudiFlight:              constant.MapStationOrbis,
			}

			boatsMovePlayers(server, arrivals)
			checkInvasion(server, true) // Remove all mobs once people have landed

			departures := map[int32]int32{
				constant.MapBoatElliniaDeparture:      constant.MapBoatElliniaFlight,
				constant.MapBoatOrbisElliniaDeparture: constant.MapBoatOrbisElliniaFlight,
				constant.MapBoatOrbisLudiDeparture:    constant.MapBoatOrbisLudiFlight,
				constant.MapBoatLudiDeparture:         constant.MapBoatLudiFlight,
			}

			boatsMovePlayers(server, departures)
			boatsBoarding(server, false)
		}

		// Note: Not sure what the spawn rate was like on GMS just remember it being low
		if r.Float64() < 0.3 {
			go func(server *Server) {
				wait(5 * time.Minute)

				finishInvasion := time.NewTimer(10 * time.Minute)
				defer finishInvasion.Stop()

				server.dispatch <- func() {
					log.Println("Boat invasion starting")
					invasion(server)
				}

				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						server.dispatch <- func() {
							checkInvasion(server, false)
						}
					case <-finishInvasion.C:
						return
					}
				}
			}(server)
		}

		wait(10 * time.Minute) // time until tickets can be bought

		server.dispatch <- func() {
			boatsBoarding(server, true)
		}
	}
}

func boatsBoarding(server *Server, canBoard bool) {
	platforms := [4]int32{
		constant.MapStationEllinia,
		constant.MapStationOrbisEllinaPlatform,
		constant.MapStationOrbisLudiPlatform,
		constant.MapStationLudiOrbisPlatform,
	}

	for _, mapID := range platforms {
		field, ok := server.fields[mapID]

		if !ok {
			log.Println("Could not allow boarding for", mapID)
			continue
		}

		for _, inst := range field.instances {
			inst.properties["canBoard"] = canBoard
			inst.showBoats(canBoard, 0x00)
		}
	}
}

func boatsMovePlayers(server *Server, warps map[int32]int32) {
	for src, dst := range warps {
		srcField, ok := server.fields[src]

		if !ok {
			log.Println("Could not not take off for", src)
			continue
		}

		dstField, ok := server.fields[dst]

		if !ok {
			log.Println("Could not not take off for", dst)
			continue
		}

		portal, err := dstField.instances[0].getPortalFromID(0)

		if err != nil {
			log.Println(err)
			continue
		}

		for _, inst := range srcField.instances {
			for _, plr := range inst.players {
				server.warpPlayer(plr, dstField, portal, true)
			}
		}
	}
}

func invasion(server *Server) {
	ships := [2]int32{constant.MapBoatElliniaFlight, constant.MapBoatOrbisElliniaFlight}

	for _, mapID := range ships {
		field, ok := server.fields[mapID]

		if !ok {
			log.Println("Could not find map", mapID)
			continue
		}

		for _, inst := range field.instances {
			inst.changeBgm("Bgm04/ArabPirate")
			inst.showBoats(true, 0x01)

			switch mapID {
			case constant.MapBoatElliniaFlight:
				inst.lifePool.spawnMobFromID(constant.MobCrimsonBalrog, newPos(485, -221, 0), false, true, true, constant.MobSummonTypeInstant, 0)
				inst.lifePool.spawnMobFromID(constant.MobCrimsonBalrog, newPos(485, -221, 0), false, true, true, constant.MobSummonTypeInstant, 0)
			case constant.MapBoatOrbisElliniaFlight:
				inst.lifePool.spawnMobFromID(constant.MobCrimsonBalrog, newPos(-590, -221, 0), false, true, true, constant.MobSummonTypeInstant, 0)
				inst.lifePool.spawnMobFromID(constant.MobCrimsonBalrog, newPos(-590, -221, 0), false, true, true, constant.MobSummonTypeInstant, 0)
			}
		}
	}
}

func checkInvasion(server *Server, finish bool) {
	ships := [2]int32{constant.MapBoatElliniaFlight, constant.MapBoatOrbisElliniaFlight}

	for _, mapID := range ships {
		field, ok := server.fields[mapID]

		if !ok {
			log.Println("Could not find map", mapID)
			continue
		}

		for _, inst := range field.instances {
			if finish {
				inst.lifePool.eraseMobs()
			}

			if len(inst.lifePool.mobs) == 0 || finish {
				inst.showBoats(false, 0x01)
				inst.changeBgm("Bgm04/UponTheSky")
			}
		}
	}
}
