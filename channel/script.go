package channel

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common/mnet"
	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
)

type scriptStore struct {
	folder   string
	scripts  map[string]*goja.Program
	dispatch chan func()
}

func createScriptStore(folder string, dispatch chan func()) *scriptStore {
	return &scriptStore{folder: folder, dispatch: dispatch, scripts: make(map[string]*goja.Program)}
}

func (s scriptStore) String() string {
	return fmt.Sprintf("%v", s.scripts)
}

func (s *scriptStore) loadScripts() error {
	err := filepath.Walk(s.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		name, program, err := createScriptProgramFromFilename(path)

		if err == nil {
			s.scripts[name] = program
		} else {
			log.Println("Script compiling:", err)
		}

		return nil
	})

	return err
}

func (s *scriptStore) monitor(task func(name string, program *goja.Program)) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Println(err)
	}

	defer watcher.Close()

	err = watcher.Add(s.folder)

	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				s.dispatch <- func() {
					log.Println("Script:", event.Name, "modified/created")
					name, program, err := createScriptProgramFromFilename(event.Name)

					if err == nil {
						s.scripts[name] = program
						task(name, program)
					} else {
						log.Println("Script compiling:", err)
					}
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				s.dispatch <- func() {
					name := filepath.Base(event.Name)
					name = strings.TrimSuffix(name, filepath.Ext(name))

					if _, ok := s.scripts[name]; ok {
						log.Println("Script:", event.Name, "removed")
						task(name, nil)
						delete(s.scripts, name)
					} else {
						log.Println("Script: could not find:", name, "to delete")
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			log.Println(err)
		}
	}
}

func createScriptProgramFromFilename(filename string) (string, *goja.Program, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", nil, err
	}

	program, err := goja.Compile(filename, string(data), false)

	if err != nil {
		return "", nil, err
	}

	filename = filepath.Base(filename)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	return name, program, nil
}

type warpFn func(plr *player, dstField *field, dstPortal portal) error

type npcScriptState struct {
	npcID       int32
	conn        mnet.Client
	terminate   bool
	selection   int32
	inputString string
	inputNumber int32

	yes, no, next, back bool // flags - covnert to bitfieds?

	goods [][]int32

	warpFunc warpFn
	fields   map[int32]*field
}

// Id of npc
func (state *npcScriptState) Id() int32 {
	return state.npcID
}

// SendBackNext packet to player
func (state *npcScriptState) SendBackNext(msg string, back, next bool) {
	state.conn.Send(packetNpcChatBackNext(state.npcID, msg, next, back))
}

// SendOK packet to player
func (state *npcScriptState) SendOK(msg string) {
	state.conn.Send(packetNpcChatOk(state.npcID, msg))
}

// SendYesNo packet to player
func (state *npcScriptState) SendYesNo(msg string) {
	state.conn.Send(packetNpcChatYesNo(state.npcID, msg))
}

// SendInputText packet to player
func (state *npcScriptState) SendInputText(msg, defaultInput string, minLength, maxLength int16) {
	state.conn.Send(packetNpcChatUserString(state.npcID, msg, defaultInput, minLength, maxLength))
}

// SendInputNumber packet to player
func (state *npcScriptState) SendInputNumber(msg string, defaultInput, minLength, maxLength int32) {
	state.conn.Send(packetNpcChatUserNumber(state.npcID, msg, defaultInput, minLength, maxLength))
}

// SendSelection packet to player
func (state *npcScriptState) SendSelection(msg string) {
	state.conn.Send(packetNpcChatSelection(state.npcID, msg))
}

// SendStyles packet to player
func (state *npcScriptState) SendStyles(msg string, styles []int32) {
	state.conn.Send(packetNpcChatStyleWindow(state.npcID, msg, styles))
}

// SendShop packet to player
func (state *npcScriptState) SendShop(goods [][]int32) {
	state.goods = goods
	state.conn.Send(packetNpcShop(state.npcID, goods))
}

// Terminate the scriopt
func (state *npcScriptState) Terminate() {
	state.terminate = true
}

// Selection value
func (state npcScriptState) Selection() int32 {
	return state.selection
}

// InputString value
func (state npcScriptState) InputString() string {
	return state.inputString
}

// InputNumber value
func (state npcScriptState) InputNumber() int32 {
	return state.inputNumber
}

// Yes flag
func (state npcScriptState) Yes() bool {
	return state.yes
}

// No flag
func (state npcScriptState) No() bool {
	return state.no
}

// Next Flag
func (state npcScriptState) Next() bool {
	return state.next
}

// Back flag
func (state npcScriptState) Back() bool {
	return state.back
}

// Goods in the shop
func (state npcScriptState) Goods() [][]int32 {
	return state.goods
}

// ClearFlags within the state
func (state *npcScriptState) ClearFlags() {
	state.next = false
	state.back = false
	state.inputNumber = -1
	state.inputString = ""
	state.selection = -1
	state.yes = false
	state.no = false
}

// SetNextBack flags
func (state *npcScriptState) SetNextBack(next, back bool) {
	state.next = next
	state.back = back
}

// SetYesNo flags
func (state *npcScriptState) SetYesNo(yes, no bool) {
	state.yes = yes
	state.no = no
}

// SetTextInput option
func (state *npcScriptState) SetTextInput(input string) {
	state.inputString = input
}

// SetNumberInput option
func (state *npcScriptState) SetNumberInput(input int32) {
	state.inputNumber = input
}

// SetOptionSelect index
func (state *npcScriptState) SetOptionSelect(selection int32) {
	state.selection = selection
}

// WarpPlayer to specific field
func (state npcScriptState) WarpPlayer(p *playerWrapper, mapID int32) bool {
	if field, ok := state.fields[mapID]; ok {
		inst, err := field.getInstance(0)

		if err != nil {
			return false
		}

		portal, err := inst.getRandomSpawnPortal()

		if err != nil {
			return false
		}

		err = state.warpFunc(p.player, field, portal)

		return err == nil
	}

	return false
}

// GetInstance that the passed in player belongs to
func (state npcScriptState) GetInstance(p *player) *fieldInstanceWrapper {
	if field, ok := state.fields[p.mapID]; ok {
		inst, err := field.getInstance(p.inst.id)

		if err != nil {
			return nil
		}

		return &fieldInstanceWrapper{inst}
	}

	return &fieldInstanceWrapper{}
}

type fieldInstanceWrapper struct {
	*fieldInstance
}

func (f *fieldInstanceWrapper) Properties(inst int) map[string]interface{} {
	if f.fieldInstance == nil {
		return make(map[string]interface{})
	}

	return f.properties
}

type npcScriptController struct {
	state npcScriptState

	vm      *goja.Runtime
	program *goja.Program

	runFunc func(*npcScriptState, *playerWrapper)
}

func createNewnpcScriptController(npcID int32, conn mnet.Client, program *goja.Program, warpFunc warpFn, fields map[int32]*field) (*npcScriptController, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, err
	}

	controller := &npcScriptController{vm: vm, program: program}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", r)
		}
	}()

	err = vm.ExportTo(vm.Get("run"), &controller.runFunc)

	if err != nil {
		return nil, err
	}

	controller.state = npcScriptState{npcID: npcID, conn: conn, warpFunc: warpFunc, fields: fields}

	return controller, nil
}

// Run the npc script
func (controller *npcScriptController) run(p *player) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", r)
		}
	}()

	controller.runFunc(&controller.state, &playerWrapper{player: p})

	return controller.state.terminate
}

type eventScriptController struct {
	name    string
	vm      *goja.Runtime
	program *goja.Program

	fields   map[int32]*field
	dispatch chan func()
	warpFunc warpFn

	initFunc  func(*eventScriptController)
	terminate bool
}

func createNewEventScriptController(name string, program *goja.Program, fields map[int32]*field, dispatch chan func(), warpFunc warpFn) (*eventScriptController, bool, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, false, err
	}

	controller := &eventScriptController{name: name, vm: vm, program: program, fields: fields, dispatch: dispatch, warpFunc: warpFunc}

	ptr := vm.Get("init")

	if ptr == nil {
		return controller, false, nil
	}

	err = vm.ExportTo(ptr, &controller.initFunc)

	if err != nil {
		return controller, false, nil
	}

	return controller, true, nil
}

// Init the system script, this is run once per event script when a controller is made for it.
// Controllers are copied for event scripts into the event manager (use init to declare global variables or schedule a function).
func (controller *eventScriptController) init() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", controller.name, r)
		}
	}()

	controller.initFunc(controller)
}

// Terminate the running script
func (controller *eventScriptController) Terminate() {
	controller.terminate = true
}

// Schedule a function in vm to run at another point
func (controller *eventScriptController) Schedule(functionName string, scheduledTime int64) {
	var scheduleFn func(*eventScriptController)

	err := controller.vm.ExportTo(controller.vm.Get(functionName), &scheduleFn)

	if err != nil {
		log.Println("Error in getting function", functionName, " from VM via scheduler in script", controller.name)
		return
	}

	go func(fnc func(*eventScriptController), scheduledTime int64, controller *eventScriptController) {
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
func (controller eventScriptController) Log(v ...interface{}) {
	log.Println(v...)
}

// ExtractFunction from program
func (controller *eventScriptController) ExtractFunction(name string, ptr *interface{}) error {

	if err := controller.vm.ExportTo(controller.vm.Get(name), ptr); err != nil {
		return err
	}

	return nil
}

// WarpPlayer to map and random spawn portal
func (controller eventScriptController) WarpPlayer(p *player, mapID int32) bool {
	if field, ok := controller.fields[mapID]; ok {
		inst, err := field.getInstance(0)

		if err != nil {
			return false
		}

		portal, err := inst.getRandomSpawnPortal()

		if err != nil {
			return false
		}

		err = controller.warpFunc(p, field, portal)

		return err == nil
	}

	return false
}

// Field wrapper
func (controller *eventScriptController) Field(id int32) *fieldWrapper {
	if field, ok := controller.fields[id]; ok {
		return &fieldWrapper{field: field, controller: controller}
	}

	return &fieldWrapper{}
}

type fieldWrapper struct {
	*field
	controller *eventScriptController
}

func (f *fieldWrapper) InstanceCount() int {
	if f.field == nil {
		return 0
	}

	return len(f.instances)
}

func (f *fieldWrapper) GetProperties(inst int) map[string]interface{} {
	if f.field == nil {
		return make(map[string]interface{})
	}

	i, err := f.getInstance(inst)

	if err != nil {
		return nil
	}

	return i.properties
}

func (f *fieldWrapper) ShowBoat(id int, show bool, boatType byte) {
	if f.field == nil {
		return
	}

	i, err := f.getInstance(id)

	if err != nil {
		return
	}

	i.showBoats(show, boatType)
}

func (f *fieldWrapper) ChangeBgm(id int, path string) {
	if f.field == nil {
		return
	}

	i, err := f.getInstance(id)

	if err != nil {
		return
	}

	i.changeBgm(path)
}

func (f *fieldWrapper) SpawnMonster(inst int, mobID int32, x, y int16, hasAgro, items, mesos bool) {
	if f.field == nil {
		return
	}

	i, err := f.getInstance(inst)

	if err != nil {
		return
	}

	_ = i.lifePool.spawnMobFromID(mobID, newPos(x, y, 0), hasAgro, items, mesos, 0)
}

func (f *fieldWrapper) Clear(id int, mobs, items bool) {
	if f.field == nil {
		return
	}

	i, err := f.getInstance(id)

	if err != nil {
		return
	}

	if mobs {
		i.lifePool.eraseMobs()
	}

	if items {
		i.dropPool.eraseDrops()
	}
}

func (f *fieldWrapper) WarpPlayersToPortal(mapID int32, portalID byte) {
	if f.field == nil {
		return
	}

	if field, ok := f.controller.fields[mapID]; ok {
		portal, err := field.instances[0].getPortalFromID(portalID)

		if err != nil {
			return
		}

		for _, i := range f.instances {
			for _, p := range i.players {
				err = f.controller.warpFunc(p, field, portal)

				if err != nil {
					return
				}
			}
		}
	}
}

func (f *fieldWrapper) MobCount(id int) int {
	if f.field == nil {
		return 0
	}

	i, err := f.getInstance(id)

	if err != nil {
		return 0
	}

	return i.lifePool.mobCount()
}

type playerWrapper struct {
	*player
}

func (p *playerWrapper) Mesos() int32 {
	return p.mesos
}

func (p *playerWrapper) GiveMesos(amount int32) {
	p.giveMesos(amount)
}
