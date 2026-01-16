package channel

import (
	"log"
	"slices"
	"sync"
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
	onMapChangeCallback      func(plr scriptPlayerWrapper, dst scriptMapWrapper)
	timeoutCallback          func(plr scriptPlayerWrapper)
	playerLeaveEventCallback func(plr scriptPlayerWrapper)

	program *goja.Program
	vm      *goja.Runtime

	closeFinish func()
	timerReset  chan struct{}
}

func createEvent(id int32, instID int, players []int32, server *Server, program *goja.Program) (*event, error) {
	ctrl := &event{
		id:         id,
		finished:   make(chan struct{}),
		instanceID: instID,
		playerIDs:  players,
		server:     server,
		program:    program,
		vm:         goja.New(),
		timerReset: make(chan struct{}, 1),
	}

	ctrl.closeFinish = sync.OnceFunc(func() {
		close(ctrl.finished)
	})

	ctrl.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	_ = ctrl.vm.Set("ctrl", ctrl)

	_, err := ctrl.vm.RunProgram(ctrl.program)

	if err != nil {
		return nil, err
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("start"), &ctrl.startCallback)

	if err != nil {
		return nil, err
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("beforePortal"), &ctrl.beforePortalCallback)

	if err != nil {
		return nil, err
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("afterPortal"), &ctrl.afterPortalCallback)

	if err != nil {
		return nil, err
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("playerLeaveEvent"), &ctrl.playerLeaveEventCallback)

	if err != nil {
		return nil, err
	}

	if fn := ctrl.vm.Get("onMapChange"); fn != nil && !goja.IsUndefined(fn) {
		_ = ctrl.vm.ExportTo(fn, &ctrl.onMapChangeCallback)
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("playerLeaveEvent"), &ctrl.playerLeaveEventCallback)

	if err != nil {
		return nil, err
	}

	err = ctrl.vm.ExportTo(ctrl.vm.Get("timeout"), &ctrl.timeoutCallback)

	if err != nil {
		return nil, err
	}

	return ctrl, nil
}

func (e *event) start() {
	e.startCallback()

	for _, id := range e.playerIDs {
		if plr, err := e.server.players.GetFromID(id); err == nil {
			plr.event = e
		}
	}

	go func() {
		timeout := time.NewTimer(e.duration)
		defer timeout.Stop()

		for {
			select {
			case <-timeout.C:
				e.server.dispatch <- func() {
					for _, id := range e.playerIDs {
						if plr, err := e.server.players.GetFromID(id); err == nil {
							plr.event = nil
							e.timeoutCallback(scriptPlayerWrapper{plr: plr, server: e.server})
						}
					}

					delete(e.server.events, e.id)
				}
				return

			case <-e.timerReset:
				if !timeout.Stop() {
					select {
					case <-timeout.C:
					default:
					}
				}
				timeout.Reset(e.duration)

			case <-e.finished:
				e.server.dispatch <- func() {
					for _, id := range e.playerIDs {
						if plr, err := e.server.players.GetFromID(id); err == nil {
							plr.event = nil
						}
					}

					delete(e.server.events, e.id)
				}
				return
			}
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
	e.closeFinish()
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

	select {
	case e.timerReset <- struct{}{}:
	default:
	}
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

func (e *event) WarpPlayers(dst int32) {
	field := e.server.fields[dst]
	dstInst, err := field.getInstance(e.instanceID)
	if err != nil {
		dstInst, err = field.getInstance(0)
		if err != nil {
			return
		}
	}

	dstPortal, err := dstInst.getRandomSpawnPortal()
	if err != nil {
		return
	}

	for _, id := range e.playerIDs {
		if plr, err := e.server.players.GetFromID(id); err != nil {
			e.server.warpPlayer(plr, field, dstPortal, false)
		}
	}
}

func (e *event) IsParticipantsOnMap(mapID int32) bool {
	for _, id := range e.playerIDs {
		if plr, err := e.server.players.GetFromID(id); err != nil {
			if plr.mapID != mapID {
				return false
			}
		}
	}
	return true
}
