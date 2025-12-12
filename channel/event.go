package channel

import (
	"log"
	"slices"
	"time"

	"github.com/dop251/goja"
)

type event struct {
	id         int32
	duration   time.Duration
	endTime    time.Time
	finished   chan struct{}
	instanceID int
	playerIDs  []int32
	server     *Server

	startCallback            func()
	beforePortalCallback     func(plr scriptPlayerWrapper, src scriptMapWrapper, dst scriptMapWrapper) bool
	afterPortalCallback      func(plr scriptPlayerWrapper, dst scriptMapWrapper)
	timeoutCallback          func(plr scriptPlayerWrapper)
	playerLeaveEventCallback func(plr scriptPlayerWrapper)

	program *goja.Program
	vm      *goja.Runtime
}

func createEvent(id int32, instID int, players []int32, server *Server, program *goja.Program) *event {
	ctrl := &event{
		id:         id,
		finished:   make(chan struct{}),
		instanceID: instID,
		playerIDs:  players,
		server:     server,
		program:    program,
		vm:         goja.New(),
	}

	ctrl.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	_ = ctrl.vm.Set("ctrl", ctrl)

	_, err := ctrl.vm.RunProgram(ctrl.program)

	if err != nil {
		log.Println(err)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("start"), &ctrl.startCallback)

	if err != nil {
		log.Println(err)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("beforePortal"), &ctrl.beforePortalCallback)

	if err != nil {
		log.Println(err)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("afterPortal"), &ctrl.afterPortalCallback)

	if err != nil {
		log.Println(err)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("playerLeaveEvent"), &ctrl.playerLeaveEventCallback)

	if err != nil {
		log.Println(err)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("timeout"), &ctrl.timeoutCallback)

	if err != nil {
		log.Println(err)
	}

	return ctrl
}

func (e *event) start(server *Server) {
	e.startCallback()

	for _, id := range e.playerIDs {
		if plr, err := e.server.players.GetFromID(id); err == nil {
			plr.event = e
		}
	}

	go func() {
		timeout := time.NewTimer(e.duration)
		defer timeout.Stop()

		select {
		case <-timeout.C:
			server.dispatch <- func() {
				for _, id := range e.playerIDs {
					if plr, err := server.players.GetFromID(id); err == nil {
						plr.event = nil
						e.timeoutCallback(scriptPlayerWrapper{plr: plr, server: e.server})
					}
				}

				delete(server.events, e.id)
			}
		case <-e.finished:
			server.dispatch <- func() {
				for _, id := range e.playerIDs {
					if plr, err := server.players.GetFromID(id); err == nil {
						plr.event = nil
					}
				}

				delete(server.events, e.id)
			}
			return
		}
	}()
}

func (e *event) Log(msg string) {
	log.Println(msg)
}

func (e *event) RemainingTime() int32 {
	return int32(time.Until(e.endTime).Seconds())
}

func (e *event) PlayerCount() int {
	return len(e.playerIDs)
}

func (e *event) Finished() {
	close(e.finished)
}

func (e *event) Players() []scriptPlayerWrapper {
	r := make([]scriptPlayerWrapper, len(e.playerIDs))

	for index, id := range e.playerIDs {
		if plr, err := e.server.players.GetFromID(id); err == nil {
			r[index] = scriptPlayerWrapper{plr: plr, server: e.server}
		}
	}

	return r
}

func (e *event) RemovePlayer(plr scriptPlayerWrapper) {
	for i, v := range e.playerIDs {
		if v == plr.plr.ID {
			e.playerIDs = slices.Delete(e.playerIDs, i, i+1)
			break
		}
	}
}

func (e *event) SetDuration(duration string) {
	countdown, err := time.ParseDuration(duration)

	if err != nil {
		countdown = time.Second * 10
	}

	e.duration = countdown
	e.endTime = time.Now().Add(countdown)
}

func (e *event) GetMap(id int32) scriptMapWrapper {
	if field, ok := e.server.fields[id]; ok {
		inst, err := field.getInstance(e.instanceID)

		if err != nil {
			return scriptMapWrapper{}
		}

		return scriptMapWrapper{inst: inst, server: e.server}
	}

	return scriptMapWrapper{}
}
