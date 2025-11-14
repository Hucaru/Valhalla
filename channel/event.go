package channel

import (
	"fmt"
	"log"
	"time"
)

type event struct {
	id           int32
	duration     time.Duration
	endTime      time.Time
	timeoutMapID int32
	finished     chan struct{}

	instanceID int

	playerIDs []int32

	// event seed

	// user portal handler - func(plr *Player, portal portal)
	// leave party handler - func(plr *Player)
	// player disconnect handler - func(plr *Player)
}

func createEvent(id int32, duration time.Duration, timeoutMapID int32, instID int, players []int32) *event {
	fmt.Println(players)
	return &event{
		id:           id,
		duration:     duration,
		timeoutMapID: timeoutMapID,
		finished:     make(chan struct{}),
		instanceID:   instID,
		playerIDs:    players,
	}
}

func (e *event) start(server *Server) {
	e.endTime = time.Now().Add(e.duration)

	go func() {
		timeout := time.NewTimer(e.duration)
		defer timeout.Stop()

		select {
		case <-timeout.C:
			server.dispatch <- func() {
				field, ok := server.fields[e.timeoutMapID]

				if !ok {
					log.Println("Could not find timeout field for event")
					return
				}

				inst, err := field.getInstance(0)

				if err != nil {
					log.Println("Could not find timeout instance for event")
					return
				}

				portal, err := inst.getRandomSpawnPortal()

				if err != nil {
					log.Println("Could not find timeout instance portal for event")
					return
				}

				for _, id := range e.playerIDs {
					if plr, err := server.players.GetFromID(id); err == nil {
						plr.eventID = 0 // TODO: change to optional type
						server.warpPlayer(plr, field, portal, true)
					}
				}

				delete(server.events, e.id)
			}
		case <-e.finished:
			server.dispatch <- func() {
				delete(server.events, e.id)
			}
			return
		}
	}()

	for _, id := range e.playerIDs {
		if plr, err := server.players.GetFromID(id); err == nil {
			plr.Send(packetShowCountdown(int32(e.duration.Seconds())))
		}
	}
}

func (e event) canWarp(mapID int32, portalName string) bool {
	return true
}

func (e *event) postWarpHook(server *Server) {
	seconds := e.endTime.Sub(time.Now()).Seconds()

	for _, id := range e.playerIDs {
		if plr, err := server.players.GetFromID(id); err == nil {
			plr.Send(packetShowCountdown(int32(seconds)))
		}
	}
}

func (e *event) leavePartyHook() {

}

func (e *event) disconnectedHook() {

}
