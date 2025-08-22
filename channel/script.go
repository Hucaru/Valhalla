package channel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common/mnet"
	"github.com/Hucaru/Valhalla/internal"
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

		name, program, errComp := createScriptProgramFromFilename(path)

		if errComp == nil {
			s.scripts[name] = program
		} else {
			log.Println("Script compiling:", errComp)
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
	data, err := os.ReadFile(filename)

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

type npcChatNodeType int

const (
	npcYesState npcChatNodeType = iota
	npcNoState
	npcNextState
	// npcBackState we don't use this as this condition is a pop rather than insert
	npcSelectionState
	npcStringInputState
	npcNumberInputState
	npcIncorrectState
)

type npcChatStateTracker struct {
	lastPos    int
	currentPos int

	list []npcChatNodeType

	selections []int32
	selection  int

	inputs []string
	input  int

	numbers []int32
	number  int
}

func (tracker *npcChatStateTracker) addState(stateType npcChatNodeType) {
	if tracker.currentPos >= len(tracker.list) {
		tracker.list = append(tracker.list, stateType)
	} else {
		tracker.list[tracker.currentPos] = stateType
	}

	tracker.currentPos++
	tracker.lastPos = tracker.currentPos
}

func (tracker *npcChatStateTracker) performInterrupt() bool {
	if tracker.currentPos == tracker.lastPos {
		return true
	}

	tracker.currentPos++

	return false
}

func (tracker *npcChatStateTracker) getCurrentState() npcChatNodeType {
	if tracker.currentPos >= len(tracker.list) {
		return npcIncorrectState
	}

	return tracker.list[tracker.currentPos]
}

func (tracker *npcChatStateTracker) popState() {
	tracker.lastPos--
}

type warpFn func(plr *player, dstField *field, dstPortal portal) error

type npcChatPlayerController struct {
	plr       *player
	fields    map[int32]*field
	warpFunc  warpFn
	worldConn mnet.Server
}

func (ctrl *npcChatPlayerController) Warp(id int32) {
	if field, ok := ctrl.fields[id]; ok {
		inst, err := field.getInstance(0)

		if err != nil {
			return
		}

		portal, err := inst.getRandomSpawnPortal()

		if err != nil {
			return
		}

		_ = ctrl.warpFunc(ctrl.plr, field, portal)
	}
}

func (ctrl *npcChatPlayerController) InstanceProperties() map[string]interface{} {
	return ctrl.plr.inst.properties
}

func (ctrl *npcChatPlayerController) Mesos() int32 {
	return ctrl.plr.mesos
}

func (ctrl *npcChatPlayerController) GiveMesos(amount int32) {
	ctrl.plr.giveMesos(amount)
}

func (ctrl *npcChatPlayerController) GiveItem(id int32, amount int16) bool {
	item, err := createItemFromID(id, amount)

	if err != nil {
		return false
	}

	if err = ctrl.plr.giveItem(item); err != nil {
		return false
	}

	return true
}

func (ctrl *npcChatPlayerController) Job() int16 {
	return ctrl.plr.job
}

func (ctrl *npcChatPlayerController) SetJob(id int16) {
	ctrl.plr.setJob(id)
}

func (ctrl *npcChatPlayerController) Level() byte {
	return ctrl.plr.level
}

func (ctrl *npcChatPlayerController) InGuild() bool {
	return ctrl.plr.guild != nil
}

func (ctrl *npcChatPlayerController) GuildRank() byte {
	if ctrl.plr.guild != nil {
		for i, id := range ctrl.plr.guild.playerID {
			if id == ctrl.plr.id {
				return ctrl.plr.guild.ranks[i]
			}
		}
	}

	return 0
}

func (ctrl *npcChatPlayerController) InParty() bool {
	return ctrl.plr.party != nil
}

func (ctrl *npcChatPlayerController) IsPartyLeader() bool {
	return ctrl.plr.party.players[0] == ctrl.plr
}

func (ctrl *npcChatPlayerController) PartyMembersOnMapCount() int {
	if ctrl.plr.party == nil {
		return 0
	}

	count := 0
	for _, v := range ctrl.plr.party.players {
		if v != nil && v.mapID == ctrl.plr.mapID {
			count++
		}
	}

	return count
}

func (ctrl *npcChatPlayerController) DisbandGuild() {
	if ctrl.plr.guild == nil {
		return
	}

	ctrl.worldConn.Send(internal.PacketGuildDisband(ctrl.plr.guild.id))
}

type npcChatController struct {
	npcID int32
	conn  mnet.Client

	goods [][]int32

	stateTracker npcChatStateTracker

	vm      *goja.Runtime
	program *goja.Program
}

func createNpcChatController(npcID int32, conn mnet.Client, program *goja.Program, plr *player, fields map[int32]*field, warpFunc warpFn, worldConn mnet.Server) (*npcChatController, error) {
	ctrl := &npcChatController{
		npcID:   npcID,
		conn:    conn,
		vm:      goja.New(),
		program: program,
	}

	plrCtrl := &npcChatPlayerController{
		plr:       plr,
		fields:    fields,
		warpFunc:  warpFunc,
		worldConn: worldConn,
	}

	ctrl.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	_ = ctrl.vm.Set("npc", ctrl)
	_ = ctrl.vm.Set("plr", plrCtrl)

	return ctrl, nil
}

func (ctrl *npcChatController) Id() int32 {
	return ctrl.npcID
}

// SendBackNext packet to player
func (ctrl *npcChatController) SendBackNext(msg string, back, next bool) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatBackNext(ctrl.npcID, msg, next, back))
		ctrl.vm.Interrupt("SendBackNext")
	}
}

// SendOK packet to player
func (ctrl *npcChatController) SendOk(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatOk(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendOk")
	}
}

// SendYesNo packet to player
func (ctrl *npcChatController) SendYesNo(msg string) bool {
	state := ctrl.stateTracker.getCurrentState()

	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatYesNo(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendYesNo")
		return false
	}

	if state == npcYesState {
		return true
	} else if state == npcNoState {
		return false
	}

	return false
}

// SendInputText packet to player
func (ctrl *npcChatController) SendInputText(msg, defaultInput string, minLength, maxLength int16) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, msg, defaultInput, minLength, maxLength))
		ctrl.vm.Interrupt("SendInputText")
	}
}

// SendInputNumber packet to player
func (ctrl *npcChatController) SendInputNumber(msg string, defaultInput, minLength, maxLength int32) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserNumber(ctrl.npcID, msg, defaultInput, minLength, maxLength))
		ctrl.vm.Interrupt("SendInputNumber")
	}
}

// SendSelection packet to player
func (ctrl *npcChatController) SendSelection(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendSelection")
	}
}

// SendStyles packet to player
func (ctrl *npcChatController) SendStyles(msg string, styles []int32) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatStyleWindow(ctrl.npcID, msg, styles))
		ctrl.vm.Interrupt("SendStyles")
	}
}

// SendGuildCreation
func (ctrl *npcChatController) SendGuildCreation() {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetGuildEnterName())
		ctrl.vm.Interrupt("SendGuildCreation")
	}
}

// SendGuildEmblemEditor
func (ctrl *npcChatController) SendGuildEmblemEditor() {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetGuildEmblemEditor())
		ctrl.vm.Interrupt("SendGuildEmblemEditor")
	}
}

// SendShop packet to player
func (ctrl *npcChatController) SendShop(goods [][]int32) {
	ctrl.goods = goods
	ctrl.conn.Send(packetNpcShop(ctrl.npcID, goods))
}

func (ctrl *npcChatController) clearUserInput() {
	ctrl.stateTracker.input = 0
	ctrl.stateTracker.selection = 0
	ctrl.stateTracker.input = 0
	ctrl.stateTracker.number = 0
}

// Selection value
func (ctrl *npcChatController) Selection() int32 {
	val := ctrl.stateTracker.selections[ctrl.stateTracker.selection]
	ctrl.stateTracker.selection++
	return val
}

// InputString value
func (ctrl *npcChatController) InputString() string {
	val := ctrl.stateTracker.inputs[ctrl.stateTracker.input]
	ctrl.stateTracker.input++
	return val
}

// InputNumber value
func (ctrl *npcChatController) InputNumber() int32 {
	val := ctrl.stateTracker.numbers[ctrl.stateTracker.number]
	ctrl.stateTracker.number++
	return val
}

func (ctrl *npcChatController) run() bool {
	ctrl.stateTracker.currentPos = 0

	_, err := ctrl.vm.RunProgram(ctrl.program)

	if _, ok := err.(*goja.InterruptedError); ok {
		return false
	}

	return true
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
	server *Server
}

func (p *playerWrapper) Mesos() int32 {
	return p.mesos
}

func (p *playerWrapper) GiveMesos(amount int32) {
	p.giveMesos(amount)
}

func (p *playerWrapper) Job() int16 {
	return p.job
}

func (p *playerWrapper) Level() int16 {
	return int16(p.level)
}

func (p *playerWrapper) GiveJob(id int16) {
	p.setJob(id)
}

func (p *playerWrapper) GainItem(id int32, amount int16) {
	item, _ := createAverageItemFromID(id, amount)
	p.giveItem(item)
}
