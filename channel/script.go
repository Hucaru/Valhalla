package channel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/internal"
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

type warpFn func(plr *Player, dstField *field, dstPortal portal) error

type npcChatPlayerController struct {
	plr       *Player
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

func (ctrl *npcChatPlayerController) WarpFromName(id int32, name string) {
	if field, ok := ctrl.fields[id]; ok {
		inst, err := field.getInstance(0)

		if err != nil {
			return
		}

		portal, err := inst.getPortalFromName(name)

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
	item, err := CreateItemFromID(id, amount)

	if err != nil {
		return false
	}

	if err = ctrl.plr.GiveItem(item); err != nil {
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
			if id == ctrl.plr.ID {
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

func (ctrl *npcChatPlayerController) GetLevel() int {
	return int(ctrl.plr.level)
}

func (ctrl *npcChatPlayerController) GetQuestStatus(id int16) int {
	// 2 = completed, 1 = in progress, 0 = not started
	for _, q := range ctrl.plr.quests.completed {
		if q.id == id {
			return 2
		}
	}
	for _, q := range ctrl.plr.quests.inProgress {
		if q.id == id {
			return 1
		}
	}
	return 0
}

func (ctrl *npcChatPlayerController) CheckQuestStatus(id int16, status int) bool {
	return ctrl.GetQuestStatus(id) == status
}

func (ctrl *npcChatPlayerController) QuestData(id int16) string {
	if q, ok := ctrl.plr.quests.inProgress[id]; ok {
		return q.name
	}
	return ""
}

func (ctrl *npcChatPlayerController) CheckQuestData(id int16, data string) bool {
	return ctrl.QuestData(id) == data
}

func (ctrl *npcChatPlayerController) SetQuestData(id int16, data string) {
	// Only allow setting data if quest is in-progress; if not, start it or ignore.
	if _, ok := ctrl.plr.quests.inProgress[id]; !ok {
		// You may choose to implicitly start; here we upsert and add in-memory if needed.
		ctrl.plr.quests.add(id, data)
	} else {
		// update in-memory
		q := ctrl.plr.quests.inProgress[id]
		q.name = data
		ctrl.plr.quests.inProgress[id] = q
	}
	// Persist + notify client
	upsertQuestRecord(ctrl.plr.ID, id, data)
	ctrl.plr.Send(packetQuestUpdate(id, data))
}

func (ctrl *npcChatPlayerController) StartQuest(id int16) bool {
	return ctrl.plr.tryStartQuest(id)
}

func (ctrl *npcChatPlayerController) CompleteQuest(id int16) bool {
	return ctrl.plr.tryCompleteQuest(id)
}

func (ctrl *npcChatPlayerController) ForfeitQuest(id int16) {
	if !ctrl.plr.quests.hasInProgress(id) {
		return
	}
	ctrl.plr.quests.remove(id)
	delete(ctrl.plr.quests.mobKills, id)
	deleteQuest(ctrl.plr.ID, id)
	ctrl.plr.Send(packetQuestRemove(id))
	clearQuestMobKills(ctrl.plr.ID, id)
}

func (ctrl *npcChatPlayerController) TakeItem(id int32, slot int16, amount int16, invID byte) bool {
	_, err := ctrl.plr.takeItem(id, slot, amount, invID)
	return err == nil
}

func (ctrl *npcChatPlayerController) RemoveItemsByID(id int32, count int32) bool {
	return ctrl.plr.removeItemsByID(id, count)
}

func (ctrl *npcChatPlayerController) ItemCount(id int32) int32 {
	return ctrl.plr.countItem(id)
}

func (ctrl *npcChatPlayerController) TakeMesos(amount int32) {
	ctrl.plr.takeMesos(amount)
}

func (ctrl *npcChatPlayerController) GetNX() int32 {
	return ctrl.plr.GetNX()
}

func (ctrl *npcChatPlayerController) SetNX(nx int32) {
	ctrl.plr.SetNX(nx)
}

func (ctrl *npcChatPlayerController) GetMaplePoints() int32 {
	return ctrl.plr.GetMaplePoints()
}

func (ctrl *npcChatPlayerController) SetMaplePoints(points int32) {
	ctrl.plr.SetMaplePoints(points)
}

func (ctrl *npcChatPlayerController) SetFame(value int16) {
	ctrl.plr.setFame(value)
}

func (ctrl *npcChatPlayerController) GiveFame(delta int16) {
	ctrl.plr.setFame(ctrl.plr.fame + delta)
}

func (ctrl *npcChatPlayerController) GiveAP(amount int16) {
	ctrl.plr.giveAP(amount)
}

func (ctrl *npcChatPlayerController) GiveSP(amount int16) {
	ctrl.plr.giveSP(amount)
}

func (ctrl *npcChatPlayerController) GiveEXP(amount int32) {
	ctrl.plr.giveEXP(amount, false, false)
}

func (ctrl *npcChatPlayerController) GiveHP(amount int16) {
	ctrl.plr.giveHP(amount)
}

func (ctrl *npcChatPlayerController) GiveMP(amount int16) {
	ctrl.plr.giveMP(amount)
}

func (ctrl *npcChatPlayerController) HealToFull() {
	ctrl.plr.setHP(ctrl.plr.maxHP)
	ctrl.plr.setMP(ctrl.plr.maxMP)
}

func (ctrl *npcChatPlayerController) Gender() byte {
	return ctrl.plr.gender
}

// Hair returns the current hair ID
func (ctrl *npcChatPlayerController) Hair() int32 {
	return ctrl.plr.hair
}

// SetHair updates the player's hair and refreshes the client appearance
func (ctrl *npcChatPlayerController) SetHair(id int32) {
	if ctrl.plr.hair == id {
		return
	}
	ctrl.plr.hair = id
	// Refresh avatar appearance on client
	ctrl.plr.Send(packetInventoryChangeEquip(*ctrl.plr))
}

// Skin returns the current skin tone (0..n)
func (ctrl *npcChatPlayerController) Skin() byte {
	return ctrl.plr.skin
}

// SetSkinColor updates the player's skin tone and refreshes the client appearance
func (ctrl *npcChatPlayerController) SetSkinColor(skin byte) {
	if ctrl.plr.skin == skin {
		return
	}
	ctrl.plr.skin = skin
	// Refresh avatar appearance on client
	ctrl.plr.Send(packetInventoryChangeEquip(*ctrl.plr))
}

type scriptQuestView struct {
	Data   string `json:"data"`
	Status int    `json:"status"` // 0,1,2 same as GetQuestStatus
}

func (ctrl *npcChatPlayerController) Quest(id int16) scriptQuestView {
	status := ctrl.GetQuestStatus(id) // already 0/1/2
	return scriptQuestView{
		Data:   ctrl.QuestData(id),
		Status: status,
	}
}

func (ctrl *npcChatPlayerController) PreviousMap() int32 {
	return ctrl.plr.previousMap
}

func (ctrl *npcChatPlayerController) MapID() int32 {
	return ctrl.plr.mapID
}

type npcChatController struct {
	npcID int32
	conn  mnet.Client

	goods       [][]int32
	persistShop bool

	stateTracker npcChatStateTracker

	vm      *goja.Runtime
	program *goja.Program
}

func createNpcChatController(npcID int32, conn mnet.Client, program *goja.Program, plr *Player, fields map[int32]*field, warpFunc warpFn, worldConn mnet.Server) (*npcChatController, error) {
	ctrl := &npcChatController{
		npcID:       npcID,
		conn:        conn,
		vm:          goja.New(),
		program:     program,
		persistShop: false,
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

// Send simple next packet to Player
func (ctrl *npcChatController) Send(text string) int {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatBackNext(ctrl.npcID, text, true, false))
		ctrl.vm.Interrupt("Send")
		return 0
	}
	if ctrl.stateTracker.getCurrentState() == npcNextState {
		return 1
	}
	return 0
}

// SendBackNext packet to Player
func (ctrl *npcChatController) SendBackNext(msg string, back, next bool) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatBackNext(ctrl.npcID, msg, next, back))
		ctrl.vm.Interrupt("SendBackNext")
	}
}

// SendOK packet to Player
func (ctrl *npcChatController) SendOk(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatOk(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendOk")
	}
}

// SendYesNo packet to Player
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

// SendInputText packet to Player
func (ctrl *npcChatController) SendInputText(msg, defaultInput string, minLength, maxLength int16) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, msg, defaultInput, minLength, maxLength))
		ctrl.vm.Interrupt("SendInputText")
	}
}

// SendInputNumber packet to Player
func (ctrl *npcChatController) SendInputNumber(msg string, defaultInput, minLength, maxLength int32) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserNumber(ctrl.npcID, msg, defaultInput, minLength, maxLength))
		ctrl.vm.Interrupt("SendInputNumber")
	}
}

// SendSelection packet to Player
func (ctrl *npcChatController) SendSelection(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendSelection")
	}
}

// SendStyles packet to Player
func (ctrl *npcChatController) SendStyles(msg string, styles []int32) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.stateTracker.addState(npcSelectionState)
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

// SendShop packet to Player
func (ctrl *npcChatController) SendShop(goods [][]int32) {
	ctrl.goods = goods
	ctrl.persistShop = true
	ctrl.conn.Send(packetNpcShop(ctrl.npcID, goods))
}

func (ctrl *npcChatController) SendMenu(baseText string, selections ...string) int {
	msg := baseText
	if len(selections) > 0 {
		var b strings.Builder
		if len(msg) > 0 {
			b.WriteString(msg)
			if msg[len(msg)-1] != '\n' {
				b.WriteByte('\n')
			}
		}
		for i, s := range selections {
			fmt.Fprintf(&b, "#L%d#%s#l\n", i, s)
		}
		msg = b.String()
	}

	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, msg))
		ctrl.vm.Interrupt("SendMenu")
		return -1
	}
	if len(ctrl.stateTracker.selections) > ctrl.stateTracker.selection {
		val := ctrl.stateTracker.selections[ctrl.stateTracker.selection]
		ctrl.stateTracker.selection++
		return int(val)
	}
	return -1
}

func (ctrl *npcChatController) SendImage(imagePath string) {
	img := fmt.Sprintf("#f%s#", imagePath)
	ctrl.SendOk(img)
}

func (ctrl *npcChatController) SendNumber(text string, def, min, max int) int {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserNumber(ctrl.npcID, text, int32(def), int32(min), int32(max)))
		ctrl.vm.Interrupt("SendNumber")
		return def
	}
	if len(ctrl.stateTracker.numbers) > ctrl.stateTracker.number {
		val := ctrl.stateTracker.numbers[ctrl.stateTracker.number]
		ctrl.stateTracker.number++
		return int(val)
	}
	return def
}

func (ctrl *npcChatController) SendBoxText(askMsg, defaultAnswer string, column, line int) string {
	max := column * line
	if max <= 0 {
		max = 200
	}
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, askMsg, defaultAnswer, 0, int16(max)))
		ctrl.vm.Interrupt("SendBoxText")
		return defaultAnswer
	}
	if len(ctrl.stateTracker.inputs) > ctrl.stateTracker.input {
		val := ctrl.stateTracker.inputs[ctrl.stateTracker.input]
		ctrl.stateTracker.input++
		return val
	}
	return defaultAnswer
}

func (ctrl *npcChatController) SendQuiz(text, problem, hint string, inputMin, inputMax, _ int) string {
	prompt := text
	if problem != "" {
		if len(prompt) > 0 {
			prompt += "\n"
		}
		prompt += problem
	}
	if hint != "" {
		prompt += "\n(" + hint + ")"
	}
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatUserString(ctrl.npcID, prompt, "", int16(inputMin), int16(inputMax)))
		ctrl.vm.Interrupt("SendQuiz")
		return ""
	}
	if len(ctrl.stateTracker.inputs) > ctrl.stateTracker.input {
		val := ctrl.stateTracker.inputs[ctrl.stateTracker.input]
		ctrl.stateTracker.input++
		return val
	}
	return ""
}

func (ctrl *npcChatController) SendSlideMenu(text string) int {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatSelection(ctrl.npcID, text))
		ctrl.vm.Interrupt("SendSlideMenu")
		return -1
	}
	if len(ctrl.stateTracker.selections) > ctrl.stateTracker.selection {
		val := ctrl.stateTracker.selections[ctrl.stateTracker.selection]
		ctrl.stateTracker.selection++
		return int(val)
	}
	return -1
}

func (ctrl *npcChatController) SendAvatar(text string, avatars ...int32) int {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.stateTracker.addState(npcSelectionState)
		ctrl.conn.Send(packetNpcChatStyleWindow(ctrl.npcID, text, avatars))
		ctrl.vm.Interrupt("SendAvatar")
		return -1
	}

	if len(ctrl.stateTracker.selections) > ctrl.stateTracker.selection {
		val := ctrl.stateTracker.selections[ctrl.stateTracker.selection]
		ctrl.stateTracker.selection++
		return int(val)
	}
	return -1
}

func (ctrl *npcChatPlayerController) InventoryExchange(itemSource int32, srcCount int32, itemExchangeFor int32, count int16) bool {
	if !ctrl.plr.removeItemsByID(itemSource, srcCount) {
		return false
	}

	item, err := CreateItemFromID(itemExchangeFor, count)
	if err != nil {
		return false
	}
	if err = ctrl.plr.GiveItem(item); err != nil {
		return false
	}
	return true
}

func (ctrl *npcChatController) clearUserInput() {
	ctrl.stateTracker.input = 0
	ctrl.stateTracker.selection = 0
	ctrl.stateTracker.input = 0
	ctrl.stateTracker.number = 0
}

// Selection value
func (ctrl *npcChatController) Selection() int32 {
	if len(ctrl.stateTracker.selections) == 0 {
		return -1
	}
	return ctrl.stateTracker.selections[len(ctrl.stateTracker.selections)-1]
}

// InputString value (non-consuming)
func (ctrl *npcChatController) InputString() string {
	if len(ctrl.stateTracker.inputs) == 0 {
		return ""
	}
	return ctrl.stateTracker.inputs[len(ctrl.stateTracker.inputs)-1]
}

// InputNumber value (non-consuming)
func (ctrl *npcChatController) InputNumber() int32 {
	if len(ctrl.stateTracker.numbers) == 0 {
		return 0
	}
	return ctrl.stateTracker.numbers[len(ctrl.stateTracker.numbers)-1]
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
func (controller eventScriptController) WarpPlayer(p *Player, mapID int32) bool {
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
	*Player
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
	p.GiveItem(item)
}

type portalScriptController struct {
	vm      *goja.Runtime
	program *goja.Program
}

type portalHost struct {
	plr    *Player
	fields map[int32]*field
	conn   mnet.Client
}

func createPortalScriptController(program *goja.Program, plr *Player, fields map[int32]*field, warpFunc warpFn, conn mnet.Client) (*portalScriptController, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	plrCtrl := &npcChatPlayerController{
		plr:       plr,
		fields:    fields,
		warpFunc:  warpFunc,
		worldConn: nil,
	}

	host := &portalHost{
		plr:    plr,
		fields: fields,
		conn:   conn,
	}

	_ = vm.Set("plr", plrCtrl)
	_ = vm.Set("portal", host)

	return &portalScriptController{
		vm:      vm,
		program: program,
	}, nil
}

func (c *portalScriptController) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in portal script:", r)
		}
	}()
	_, _ = c.vm.RunProgram(c.program)
}
