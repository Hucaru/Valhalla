package script

import (
	"fmt"
	"log"
	"time"

	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/player"
	"github.com/Hucaru/Valhalla/server/pos"
	"github.com/dop251/goja"
)

// EventController is the controller for scripts that are responsible for certain parts of the game e.g. boat rides
type EventController struct {
	name    string
	vm      *goja.Runtime
	program *goja.Program

	fields   map[int32]*field.Field
	dispatch chan func()
	warpFunc warpFn

	initFunc  func(*EventController)
	terminate bool
}

// CreateNewEventController for a specific system
func CreateNewEventController(name string, program *goja.Program, fields map[int32]*field.Field, dispatch chan func(), warpFunc warpFn) (*EventController, bool, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, false, err
	}

	controller := &EventController{name: name, vm: vm, program: program, fields: fields, dispatch: dispatch, warpFunc: warpFunc}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("No run function")
		}
	}()

	err = vm.ExportTo(vm.Get("init"), &controller.initFunc)

	if err != nil {
		return controller, false, nil
	}

	return controller, true, nil
}

// Init the system script, this is run once per event script when a controller is made for it.
// Controllers are copied for event scripts into the event manager (use init to declare global variables or schedule a function).
func (controller *EventController) Init() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", controller.name, r)
		}
	}()

	controller.initFunc(controller)
}

// Terminate the running script
func (controller *EventController) Terminate() {
	controller.terminate = true
}

// Schedule a function in vm to run at another point
func (controller *EventController) Schedule(functionName string, scheduledTime int64) {
	var scheduleFn func(*EventController)

	err := controller.vm.ExportTo(controller.vm.Get(functionName), &scheduleFn)

	if err != nil {
		log.Println("Error in getting function", functionName, " from VM via scheduler in script", controller.name)
		return
	}

	go func(fnc func(*EventController), scheduledTime int64, controller *EventController) {
		timer := time.NewTimer(time.Duration(scheduledTime) * time.Millisecond)
		<-timer.C

		if controller.terminate {
			log.Println("Script:", controller.name, "has been terminated")
			return
		}

		controller.dispatch <- func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Error in running script:", controller.name, r)
				}
			}()

			fnc(controller)
		}

	}(scheduleFn, scheduledTime, controller)
}

// Log that is safe to use by script
func (controller EventController) Log(v ...interface{}) {
	log.Println(v...)
}

// ExtractFunction from program
func (controller *EventController) ExtractFunction(name string, ptr *interface{}) error {

	if err := controller.vm.ExportTo(controller.vm.Get(name), ptr); err != nil {
		return err
	}

	return nil
}

// WarpPlayer to map and random spawn portal
func (controller EventController) WarpPlayer(p *player.Data, mapID int32) bool {
	if field, ok := controller.fields[mapID]; ok {
		inst, err := field.GetInstance(0)

		if err != nil {
			return false
		}

		portal, err := inst.GetRandomSpawnPortal()

		controller.warpFunc(p, field, portal)

		return true
	}

	return false
}

// WarpPlayerToPortal in map
func (controller EventController) WarpPlayerToPortal(p *player.Data, mapID int32, portalID byte) bool {
	if field, ok := controller.fields[mapID]; ok {
		inst, err := field.GetInstance(0)

		if err != nil {
			return false
		}

		portal, err := inst.GetPortalFromID(portalID)

		controller.warpFunc(p, field, portal)

		return true
	}

	return false
}

// Fields in the game
func (controller *EventController) Fields() map[int32]*field.Field {
	return controller.fields
}

// CreatePos from x,y co-ords
func (controller *EventController) CreatePos(x int16, y int16) pos.Data {
	return pos.New(x, y, 0)
}
