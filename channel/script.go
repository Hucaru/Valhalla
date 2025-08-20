package channel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
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

func (s *scriptStore) get(name string) (*goja.Program, bool) {
	program, ok := s.scripts[name]
	return program, ok
}

func (s *scriptStore) loadScripts() error {
	err := filepath.Walk(s.folder, func(path string, info os.FileInfo, err error) error {
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
	plr      *player
	fields   map[int32]*field
	warpFunc warpFn
}

func (ctrl *npcChatPlayerController) Warp(id int32) {
	if field, ok := ctrl.fields[id]; ok {
		inst, err := field.getInstance(0)

		if err != nil {
			return
		}

		portal, err := inst.getRandomSpawnPortal()

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

func (ctrl *npcChatPlayerController) TakeMesos(amount int32) {
	ctrl.plr.takeMesos(amount)
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

func (ctrl *npcChatPlayerController) SetLevel(level byte) {
	ctrl.plr.setLevel(level)
}

func (ctrl *npcChatPlayerController) TakeItem(id int32, amount int16) bool {
	item, err := ctrl.plr.takeItemAnySlot(id, amount)
	if err != nil {
		return false
	}

	_, err = ctrl.plr.takeItem(item.id, item.slotID, item.amount, item.invID)
	if err != nil {
		return false
	}

	return true
}

func (ctrl *npcChatPlayerController) HasEquipped(id int32) bool {
	return ctrl.plr.hasEquipped(id)
}

func (ctrl *npcChatPlayerController) SetSkinColor(skin byte) {
	ctrl.plr.setSkinColor(skin)
}

func (ctrl *npcChatPlayerController) MapId() int32 {
	return ctrl.plr.mapID
}

type npcChatController struct {
	npcID int32
	conn  mnet.Client

	lastSelection   int32
	lastInputString string
	lastInputNumber int32

	goods [][]int32

	stateTracker npcChatStateTracker

	vm      *goja.Runtime
	program *goja.Program
}

func createNpcChatController(npcID int32, conn mnet.Client, program *goja.Program, plr *player, fields map[int32]*field, warpFunc warpFn) (*npcChatController, error) {
	ctrl := &npcChatController{
		npcID:   npcID,
		conn:    conn,
		vm:      goja.New(),
		program: program,
	}

	plrCtrl := &npcChatPlayerController{
		plr:      plr,
		fields:   fields,
		warpFunc: warpFunc,
	}

	ctrl.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	ctrl.vm.Set("npc", ctrl)
	ctrl.vm.Set("plr", plrCtrl)

	return ctrl, nil
}

// Id of npc
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

// AskMenu opens a selection window.
func (ctrl *npcChatController) AskMenu(baseText string, selections ...string) int32 {
	if len(selections) == 0 {
		// Treat baseText as already formatted selection list
		if ctrl.stateTracker.performInterrupt() {
			ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, baseText))
			ctrl.vm.Interrupt("AskMenu")
			return 0
		}
		return ctrl.lastSelection
	}

	// Build selection text as 1-based options
	var b strings.Builder
	if baseText != "" {
		b.WriteString(baseText)
		if !strings.HasSuffix(baseText, "\r\n") {
			b.WriteString("\r\n")
		}
	}
	for i, opt := range selections {
		// Options numbered 1..N to match typical cm.askMenu usage
		fmt.Fprintf(&b, "#L%d#%s#l\r\n", i+1, opt)
	}

	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, b.String()))
		ctrl.vm.Interrupt("AskMenu")
		return 0
	}
	return ctrl.lastSelection
}

// AskSlideMenu emulates a slide menu using a standard selection window.
func (ctrl *npcChatController) AskSlideMenu(text string) int32 {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, text))
		ctrl.vm.Interrupt("AskSlideMenu")
		return 0
	}
	return ctrl.lastSelection
}

// AskAvatar opens the style chooser and returns the chosen style id/index (engine will provide via selection).
func (ctrl *npcChatController) AskAvatar(text string, avatars ...int32) int32 {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatStyleWindow(ctrl.npcID, text, avatars))
		ctrl.vm.Interrupt("AskAvatar")
		return 0
	}
	return ctrl.lastSelection
}

// AskImage shows an image prompt using a Back/Next page and returns a simple OK(1)/Cancel(0) style selection.
func (ctrl *npcChatController) AskImage(imagePath string, extraData int32) int32 {
	// Compose a selection that includes the image and two choices
	msg := fmt.Sprintf("#F%s#\r\n#L1#OK#l\r\n#L0#Cancel#l\r\n", imagePath)
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, msg))
		ctrl.vm.Interrupt("AskImage")
		return 0
	}
	return ctrl.lastSelection
}

// AskText prompts for free text (shortcut wrapper around SendInputText).
func (ctrl *npcChatController) AskText(text string) string {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, text, "", 0, 300))
		ctrl.vm.Interrupt("AskText")
		return ""
	}
	return ctrl.lastInputString
}

// AskBoxText prompts for larger text with a suggested size (column*line). Falls back to a safe max if <= 0.
func (ctrl *npcChatController) AskBoxText(askMsg string, defaultAnswer string, column, line int) string {
	maxLen := int16(column * line)
	if maxLen <= 0 {
		maxLen = 600
	}
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, askMsg, defaultAnswer, 0, maxLen))
		ctrl.vm.Interrupt("AskBoxText")
		return ""
	}
	return ctrl.lastInputString
}

// AskQuiz emulates a quiz input by using a text input field. Time limits are not enforced at engine-level here.
func (ctrl *npcChatController) AskQuiz(text, problem, hint string, inputMin, inputMax, limitTime int) string {
	// Combine prompt + problem + optional hint
	var b strings.Builder
	if text != "" {
		b.WriteString(text)
		b.WriteString("\r\n")
	}
	if problem != "" {
		b.WriteString(problem)
		b.WriteString("\r\n")
	}
	if hint != "" {
		b.WriteString("#g")
		b.WriteString(hint)
		b.WriteString("#k\r\n")
	}
	maxLen := int16(inputMax)
	if maxLen <= 0 {
		maxLen = 300
	}
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, b.String(), "", int16(inputMin), maxLen))
		ctrl.vm.Interrupt("AskQuiz")
		return ""
	}
	return ctrl.lastInputString
}

// AskNumber prompts for a number with bounds (shortcut wrapper around SendInputNumber).
func (ctrl *npcChatController) AskNumber(text string, def, min, max int32) int32 {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserNumber(ctrl.npcID, text, def, min, max))
		ctrl.vm.Interrupt("AskNumber")
		return 0
	}
	return ctrl.lastInputNumber
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

// SendShop packet to player
func (ctrl *npcChatController) SendShop(goods [][]int32) {
	ctrl.goods = goods
	ctrl.conn.Send(packetNpcShop(ctrl.npcID, goods))
}

func (ctrl *npcChatController) clearUserInput() {
	ctrl.lastSelection = 0
	ctrl.lastInputString = ""
	ctrl.lastInputNumber = 0
}

// Selection value
func (ctrl *npcChatController) Selection() int32 {
	return ctrl.lastSelection
}

// InputString value
func (ctrl *npcChatController) InputString() string {
	return ctrl.lastInputString
}

// InputNumber value
func (ctrl *npcChatController) InputNumber() int32 {
	return ctrl.lastInputNumber
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

		controller.warpFunc(p, field, portal)

		return true
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

	i.lifePool.spawnMobFromID(mobID, newPos(x, y, 0), hasAgro, items, mesos)
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
				f.controller.warpFunc(p, field, portal)
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

func (p *playerWrapper) TakeItem(id int32, amount int16) {
	p.takeItemAnySlot(id, amount)
}
