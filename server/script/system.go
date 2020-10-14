package script

import (
	"fmt"
	"log"
	"time"

	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/player"
	"github.com/dop251/goja"
)

// SystemController is the controller for scripts that are responsible for certain parts of the game e.g. boat rides
type SystemController struct {
	name    string
	vm      *goja.Runtime
	program *goja.Program

	fields    map[int32]*field.Field
	dispatch  chan func()
	runFunc   func(*SystemController)
	terminate bool
}

// CreateNewSystemController for a specific system
func CreateNewSystemController(name string, program *goja.Program, fields map[int32]*field.Field, dispatch chan func()) (*SystemController, bool, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, false, err
	}

	controller := &SystemController{name: name, vm: vm, program: program, fields: fields, dispatch: dispatch}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("No run function")
		}
	}()

	err = vm.ExportTo(vm.Get("run"), &controller.runFunc)

	if err != nil {
		return controller, false, nil
	}

	return controller, true, nil
}

// Start the system script
func (controller *SystemController) Start() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", controller.name, r)
		}
	}()

	controller.runFunc(controller)
}

// Terminate the running script
func (controller *SystemController) Terminate() {
	controller.terminate = true
}

// Schedule a function in vm to run at another point
func (controller *SystemController) Schedule(functionName string, scheduledTime int64) {
	var scheduleFn func(*SystemController)

	err := controller.vm.ExportTo(controller.vm.Get(functionName), &scheduleFn)

	if err != nil {
		log.Println("Error in getting function", functionName, " from VM via scheduler in script", controller.name)
		return
	}

	go func(fnc func(*SystemController), scheduledTime int64, controller *SystemController) {
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
func (controller SystemController) Log(v ...interface{}) {
	log.Println(v...)
}

// WarpPlayer to map and random spawn portal
func (controller SystemController) WarpPlayer(p *player.Data, mapID int32) bool {
	return true
}

// WarpPlayerToPortal in map
func (controller SystemController) WarpPlayerToPortal(p *player.Data, mapID int32, portalID byte) bool {
	return true
}

// Fields in the game
func (controller *SystemController) Fields() map[int32]*field.Field {
	return controller.fields
}
