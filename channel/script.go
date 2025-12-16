package channel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
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

type scriptPlayerWrapper struct {
	plr    *Player
	server *Server
}

func (ctrl *scriptPlayerWrapper) Warp(id int32) {
	if field, ok := ctrl.server.fields[id]; ok {
		inst, err := field.getInstance(ctrl.plr.inst.id)

		if err != nil {
			inst, err = field.getInstance(0)

			if err != nil {
				return
			}
		}

		portal, err := inst.getRandomSpawnPortal()

		if err != nil {
			return
		}

		ctrl.server.warpPlayer(ctrl.plr, field, portal, true)
	}
}

func (ctrl *scriptPlayerWrapper) SendMessage(msg string) {
	ctrl.plr.Send(packetMessageRedText(msg))
}

func (ctrl *scriptPlayerWrapper) Mesos() int32 {
	return ctrl.plr.mesos
}

func (ctrl *scriptPlayerWrapper) GiveMesos(amount int32) {
	ctrl.plr.giveMesos(amount)
}

func (ctrl *scriptPlayerWrapper) GiveItem(id int32, amount int16) bool {
	item, err := CreateItemFromID(id, amount)

	if err != nil {
		return false
	}

	if err, _ = ctrl.plr.GiveItem(item); err != nil {
		return false
	}

	return true
}

func (ctrl *scriptPlayerWrapper) Job() int16 {
	return ctrl.plr.job
}

func (ctrl *scriptPlayerWrapper) SetJob(id int16) {
	ctrl.plr.setJob(id)
}

func (ctrl *scriptPlayerWrapper) Level() byte {
	return ctrl.plr.level
}

func (ctrl *scriptPlayerWrapper) InGuild() bool {
	return ctrl.plr.guild != nil
}

func (ctrl *scriptPlayerWrapper) GuildRank() byte {
	if ctrl.plr.guild != nil {
		for i, id := range ctrl.plr.guild.playerID {
			if id == ctrl.plr.ID {
				return ctrl.plr.guild.ranks[i]
			}
		}
	}

	return 0
}

func (ctrl *scriptPlayerWrapper) InParty() bool {
	return ctrl.plr.party != nil
}

func (ctrl *scriptPlayerWrapper) IsPartyLeader() bool {
	if ctrl.InParty() {
		return ctrl.plr.party.players[0] == ctrl.plr
	}

	return false
}

func (ctrl *scriptPlayerWrapper) PartyMembersOnMapCount() int {
	if !ctrl.InParty() {
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

func (ctrl *scriptPlayerWrapper) PartyMembersOnMap() []scriptPlayerWrapper {
	if !ctrl.InParty() {
		return []scriptPlayerWrapper{}
	}

	members := make([]scriptPlayerWrapper, 0, constant.MaxPartySize)

	for _, v := range ctrl.plr.party.players {
		if v != nil {
			members = append(members, scriptPlayerWrapper{plr: v, server: ctrl.server})
		}
	}

	return members
}

func (ctrl *scriptPlayerWrapper) PartyGiveExp(val int32) {
	if !ctrl.InParty() {
		return
	}

	for _, plr := range ctrl.plr.party.players {
		if plr != nil {
			plr.giveEXP(val, false, false)
		}
	}
}

func (ctrl *scriptPlayerWrapper) DisbandGuild() {
	if ctrl.plr.guild == nil {
		return
	}

	ctrl.server.world.Send(internal.PacketGuildDisband(ctrl.plr.guild.id))
}

func (ctrl *scriptPlayerWrapper) GetLevel() int {
	return int(ctrl.plr.level)
}

func (ctrl *scriptPlayerWrapper) GetQuestStatus(id int16) int {
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

func (ctrl *scriptPlayerWrapper) CheckQuestStatus(id int16, status int) bool {
	return ctrl.GetQuestStatus(id) == status
}

func (ctrl *scriptPlayerWrapper) QuestData(id int16) string {
	if q, ok := ctrl.plr.quests.inProgress[id]; ok {
		return q.name
	}
	return ""
}

func (ctrl *scriptPlayerWrapper) CheckQuestData(id int16, data string) bool {
	return ctrl.QuestData(id) == data
}

func (ctrl *scriptPlayerWrapper) SetQuestData(id int16, data string) {
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

func (ctrl *scriptPlayerWrapper) StartQuest(id int16) bool {
	return ctrl.plr.tryStartQuest(id)
}

func (ctrl *scriptPlayerWrapper) CompleteQuest(id int16) bool {
	return ctrl.plr.tryCompleteQuest(id)
}

func (ctrl *scriptPlayerWrapper) ForfeitQuest(id int16) {
	if !ctrl.plr.quests.hasInProgress(id) {
		return
	}
	ctrl.plr.quests.remove(id)
	delete(ctrl.plr.quests.mobKills, id)
	deleteQuest(ctrl.plr.ID, id)
	ctrl.plr.Send(packetQuestRemove(id))
	clearQuestMobKills(ctrl.plr.ID, id)
}

func (ctrl *scriptPlayerWrapper) TakeItem(id int32, slot int16, amount int16, invID byte) bool {
	_, err := ctrl.plr.takeItem(id, slot, amount, invID)
	return err == nil
}

func (ctrl *scriptPlayerWrapper) RemoveItemsByID(id int32, count int32) bool {
	return ctrl.plr.removeItemsByID(id, count)
}

func (ctrl *scriptPlayerWrapper) ItemCount(id int32) int32 {
	return ctrl.plr.countItem(id)
}

func (ctrl *scriptPlayerWrapper) TakeMesos(amount int32) {
	ctrl.plr.takeMesos(amount)
}

func (ctrl *scriptPlayerWrapper) GetNX() int32 {
	return ctrl.plr.GetNX()
}

func (ctrl *scriptPlayerWrapper) SetNX(nx int32) {
	ctrl.plr.SetNX(nx)
}

func (ctrl *scriptPlayerWrapper) GetMaplePoints() int32 {
	return ctrl.plr.GetMaplePoints()
}

func (ctrl *scriptPlayerWrapper) SetMaplePoints(points int32) {
	ctrl.plr.SetMaplePoints(points)
}

func (ctrl *scriptPlayerWrapper) SetFame(value int16) {
	ctrl.plr.setFame(value)
}

func (ctrl *scriptPlayerWrapper) GiveFame(delta int16) {
	ctrl.plr.setFame(ctrl.plr.fame + delta)
}

func (ctrl *scriptPlayerWrapper) GiveAP(amount int16) {
	ctrl.plr.giveAP(amount)
}

func (ctrl *scriptPlayerWrapper) GiveSP(amount int16) {
	ctrl.plr.giveSP(amount)
}

func (ctrl *scriptPlayerWrapper) GiveEXP(amount int32) {
	ctrl.plr.giveEXP(amount, false, false)
}

func (ctrl *scriptPlayerWrapper) GiveHP(amount int16) {
	ctrl.plr.giveHP(amount)
}

func (ctrl *scriptPlayerWrapper) GiveMP(amount int16) {
	ctrl.plr.giveMP(amount)
}

func (ctrl *scriptPlayerWrapper) HealToFull() {
	ctrl.plr.setHP(ctrl.plr.maxHP)
	ctrl.plr.setMP(ctrl.plr.maxMP)
}

func (ctrl *scriptPlayerWrapper) Gender() byte {
	return ctrl.plr.gender
}

// Hair returns the current hair ID
func (ctrl *scriptPlayerWrapper) Hair() int32 {
	return ctrl.plr.hair
}

// SetHair updates the player's hair and refreshes the client appearance
func (ctrl *scriptPlayerWrapper) SetHair(id int32) {
	if ctrl.plr.hair == id {
		return
	}
	err := ctrl.plr.setHair(id)
	if err != nil {
		return
	}
}

func (ctrl *scriptPlayerWrapper) Face() int32 {
	return ctrl.plr.face
}

func (ctrl *scriptPlayerWrapper) SetFace(id int32) {
	if ctrl.plr.face == id {
		return
	}
	err := ctrl.plr.setFace(id)
	if err != nil {
		return
	}
}

// Skin returns the current skin tone (0..n)
func (ctrl *scriptPlayerWrapper) Skin() byte {
	return ctrl.plr.skin
}

// SetSkinColor updates the player's skin tone and refreshes the client appearance
func (ctrl *scriptPlayerWrapper) SetSkinColor(skin byte) {
	if ctrl.plr.skin == skin {
		return
	}
	err := ctrl.plr.setSkin(skin)
	if err != nil {
		return
	}
}

type scriptQuestView struct {
	Data   string `json:"data"`
	Status int    `json:"status"` // 0,1,2 same as GetQuestStatus
}

func (ctrl *scriptPlayerWrapper) Quest(id int16) scriptQuestView {
	status := ctrl.GetQuestStatus(id) // already 0/1/2
	return scriptQuestView{
		Data:   ctrl.QuestData(id),
		Status: status,
	}
}

func (ctrl *scriptPlayerWrapper) PreviousMap() int32 {
	return ctrl.plr.previousMap
}

func (ctrl *scriptPlayerWrapper) MapID() int32 {
	return ctrl.plr.mapID
}

func (ctrl *scriptPlayerWrapper) Position() map[string]int16 {
	return map[string]int16{
		"x": ctrl.plr.pos.x,
		"y": ctrl.plr.pos.y,
	}
}

func (ctrl *scriptPlayerWrapper) Name() string {
	return ctrl.plr.Name
}

func (ctrl *scriptPlayerWrapper) InventoryExchange(itemSource int32, srcCount int32, itemExchangeFor int32, count int16) bool {
	if !ctrl.plr.removeItemsByID(itemSource, srcCount) {
		return false
	}

	item, err := CreateItemFromID(itemExchangeFor, count)
	if err != nil {
		return false
	}
	if err, _ = ctrl.plr.GiveItem(item); err != nil {
		return false
	}
	return true
}

func (ctrl *scriptPlayerWrapper) ShowCountdown(seconds int32) {
	ctrl.plr.Send(packetShowCountdown(seconds))
}

func (ctrl *scriptPlayerWrapper) PortalEffect(path string) {
	ctrl.plr.Send(packetPortalEffectt(2, path))
}

func (ctrl *scriptPlayerWrapper) StartPartyQuest(name string, instID int) {
	if ctrl.plr.party == nil {
		return
	}

	program, ok := ctrl.server.eventScriptStore.scripts[name]

	if !ok {
		return
	}

	ids := []int32{}

	if ctrl.plr.party != nil {
		for i, id := range ctrl.plr.party.PlayerID {
			if ctrl.plr.mapID == ctrl.plr.party.MapID[i] && ctrl.plr.party.players[i] != nil {
				if ctrl.plr.inst.id == ctrl.plr.party.players[i].inst.id {
					ids = append(ids, id)
				}
			}
		}
	} else {
		ids = append(ids, ctrl.plr.ID)
	}

	event, err := createEvent(ctrl.plr.ID, instID, ids, ctrl.server, program)

	if err != nil {
		log.Println(err)
		return
	}

	ctrl.server.events[ctrl.plr.party.ID] = event
	event.start()
}

func (ctrl *scriptPlayerWrapper) LeavePartyQuest() {
	if ctrl.plr.party == nil {
		return
	}

	if event, ok := ctrl.server.events[ctrl.plr.party.ID]; ok {
		event.playerLeaveEventCallback(scriptPlayerWrapper{plr: ctrl.plr, server: ctrl.server})
	}
}

type scriptMapWrapper struct {
	inst   *fieldInstance
	server *Server
}

func (ctrl *scriptMapWrapper) PlayerCount(mapID int32, instID int) int {
	f, ok := ctrl.server.fields[mapID]

	if !ok {
		return 0
	}

	inst, err := f.getInstance(instID)

	if err != nil {
		return 0
	}

	return len(inst.players)
}

func (ctrl *scriptMapWrapper) PlaySound(path string) {
	ctrl.inst.send(packetPlaySound(path))
}

func (ctrl *scriptMapWrapper) ShowEffect(path string) {
	ctrl.inst.send(packetShowScreenEffect(path))
}

func (ctrl *scriptMapWrapper) PortalEffect(path string) {
	ctrl.inst.send(packetPortalEffectt(2, path))
}

func (ctrl *scriptMapWrapper) Properties() map[string]interface{} {
	return ctrl.inst.properties
}

func (ctrl *scriptMapWrapper) ClearProperties() {
	for k := range ctrl.inst.properties {
		delete(ctrl.inst.properties, k)
	}
}

func (ctrl *scriptMapWrapper) PlayersInArea(id int) int {
	areas := nx.GetMaps()[ctrl.inst.fieldID].Areas
	count := 0

	for _, plr := range ctrl.inst.players {
		if areas[id].Inside(plr.pos.x, plr.pos.y) {
			count++
		}

	}

	return count
}

func (ctrl *scriptMapWrapper) MobCount() int {
	return ctrl.inst.lifePool.mobCount()
}

func (ctrl *scriptMapWrapper) RemoveDrops() {
	ctrl.inst.dropPool.clearDrops()
}

func (ctrl *scriptMapWrapper) GetMap(id int32, instID int) scriptMapWrapper {
	if field, ok := ctrl.server.fields[id]; ok {
		inst, err := field.getInstance(instID)

		if err != nil {
			instID = field.createInstance(&ctrl.server.rates, ctrl.server)
			inst, err = field.getInstance(instID)

			if err != nil {
				return scriptMapWrapper{}
			}

			return scriptMapWrapper{inst: inst, server: ctrl.server}
		}

		return scriptMapWrapper{inst: inst, server: ctrl.server}
	}

	return scriptMapWrapper{}
}

type npcChatController struct {
	npcID int32
	conn  mnet.Client

	goods [][]int32

	stateTracker npcChatStateTracker

	vm      *goja.Runtime
	program *goja.Program

	selectionCalls int
}

func createNpcChatController(npcID int32, conn mnet.Client, program *goja.Program, plr *Player, server *Server) (*npcChatController, error) {
	ctrl := &npcChatController{
		npcID:   npcID,
		conn:    conn,
		vm:      goja.New(),
		program: program,
	}

	plrCtrl := &scriptPlayerWrapper{
		plr:    plr,
		server: server,
	}

	mapWrapper := &scriptMapWrapper{
		inst:   plr.inst,
		server: server,
	}

	ctrl.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	_ = ctrl.vm.Set("npc", ctrl)
	_ = ctrl.vm.Set("plr", plrCtrl)
	_ = ctrl.vm.Set("map", mapWrapper)

	return ctrl, nil
}

func (ctrl *npcChatController) Id() int32 {
	return ctrl.npcID
}

// SendNext simple next packet to Player
func (ctrl *npcChatController) SendNext(text string) int {
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
func (ctrl *npcChatController) SendBackNext(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatBackNext(ctrl.npcID, msg, true, true))
		ctrl.vm.Interrupt("SendBackNext")
	}
}

// SendBackNext packet to Player
func (ctrl *npcChatController) SendBack(msg string) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.conn.Send(packetNpcChatBackNext(ctrl.npcID, msg, false, true))
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

func (ctrl *npcChatController) SendAvatar(text string, avatars ...int32) {
	if ctrl.stateTracker.performInterrupt() {
		ctrl.stateTracker.addState(npcSelectionState)
		ctrl.conn.Send(packetNpcChatStyleWindow(ctrl.npcID, text, avatars))
		ctrl.vm.Interrupt("SendAvatar")
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
	if ctrl.stateTracker.performInterrupt() {
		ctrl.goods = goods
		ctrl.conn.Send(packetNpcShop(ctrl.npcID, goods))
		ctrl.vm.Interrupt("SendShop")
	}
}

func (ctrl *npcChatController) SendStorage(npcID int32) {
	if ctrl.stateTracker.performInterrupt() {
		var storageMesos int32
		var storageSlots byte
		var allItems []Item

		accountID := ctrl.conn.GetAccountID()
		if accountID != 0 {
			st := new(storage)
			if err := st.load(accountID); err == nil {
				storageMesos = st.mesos
				storageSlots = st.maxSlots
				allItems = st.getAllItems()
			}
		}

		ctrl.conn.Send(packetNpcStorageShow(npcID, storageMesos, storageSlots, allItems))
		ctrl.vm.Interrupt("SendStorage")
	}
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

func (ctrl *npcChatController) clearUserInput() {
	// Reset counters but preserve the data
	ctrl.stateTracker.selection = 0
	ctrl.stateTracker.input = 0
	ctrl.stateTracker.number = 0
}

// Selection value
func (ctrl *npcChatController) Selection() int32 {
	if len(ctrl.stateTracker.selections) == 0 {
		return -1
	}
	if ctrl.stateTracker.selection >= len(ctrl.stateTracker.selections) {
		return ctrl.stateTracker.selections[len(ctrl.stateTracker.selections)-1]
	}
	val := ctrl.stateTracker.selections[ctrl.stateTracker.selection]
	ctrl.stateTracker.selection++
	return val
}

func (ctrl *npcChatController) InputString() string {
	if len(ctrl.stateTracker.inputs) == 0 {
		return ""
	}
	if ctrl.stateTracker.input >= len(ctrl.stateTracker.inputs) {
		return ctrl.stateTracker.inputs[len(ctrl.stateTracker.inputs)-1]
	}
	val := ctrl.stateTracker.inputs[ctrl.stateTracker.input]
	ctrl.stateTracker.input++
	return val
}

func (ctrl *npcChatController) InputNumber() int32 {
	if len(ctrl.stateTracker.numbers) == 0 {
		return 0
	}
	if ctrl.stateTracker.number >= len(ctrl.stateTracker.numbers) {
		return ctrl.stateTracker.numbers[len(ctrl.stateTracker.numbers)-1]
	}
	val := ctrl.stateTracker.numbers[ctrl.stateTracker.number]
	ctrl.stateTracker.number++
	return val
}

func (ctrl *npcChatController) run() bool {
	currentConversationPos := ctrl.stateTracker.currentPos
	ctrl.selectionCalls = 0

	if currentConversationPos == 0 && ctrl.stateTracker.lastPos == 0 {
		ctrl.stateTracker.selections = ctrl.stateTracker.selections[:0]
	} else {
		ctrl.stateTracker.currentPos = 0
	}

	if ctrl.vm == nil || ctrl.program == nil {
		return true
	}

	_, err := ctrl.vm.RunProgram(ctrl.program)

	if err != nil {
		if _, isInterrupted := err.(*goja.InterruptedError); isInterrupted {
			return false
		}
		return true
	}

	if ctrl.stateTracker.currentPos >= ctrl.stateTracker.lastPos {
		return true
	}
	return false
}
