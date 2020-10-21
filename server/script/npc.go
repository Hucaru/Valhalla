package script

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/player"
	"github.com/dop251/goja"
)

type warpFn func(plr *player.Data, dstField *field.Field, dstPortal field.Portal) error

// unexported as we only want the controller to make these
type npcState struct {
	npcID       int32
	conn        mnet.Client
	terminate   bool
	selection   int32
	inputString string
	inputNumber int32

	yes, no, next, back bool // flags - covnert to bitfieds?

	goods [][]int32

	warpFunc warpFn
	fields   map[int32]*field.Field
}

// SendBackNext packet to player
func (state *npcState) SendBackNext(msg string, back, next bool) {
	state.conn.Send(packetChatBackNext(state.npcID, msg, next, back))
}

// SendOK packet to player
func (state *npcState) SendOK(msg string) {
	state.conn.Send(packetChatOk(state.npcID, msg))
}

// SendYesNo packet to player
func (state *npcState) SendYesNo(msg string) {
	state.conn.Send(packetChatYesNo(state.npcID, msg))
}

// SendInputText packet to player
func (state *npcState) SendInputText(msg, defaultInput string, minLength, maxLength int16) {
	state.conn.Send(packetChatUserString(state.npcID, msg, defaultInput, minLength, maxLength))
}

// SendInputNumber packet to player
func (state *npcState) SendInputNumber(msg string, defaultInput, minLength, maxLength int32) {
	state.conn.Send(packetChatUserNumber(state.npcID, msg, defaultInput, minLength, maxLength))
}

// SendSelection packet to player
func (state *npcState) SendSelection(msg string) {
	state.conn.Send(packetChatSelection(state.npcID, msg))
}

// SendStyles packet to player
func (state *npcState) SendStyles(msg string, styles []int32) {
	state.conn.Send(packetChatStyleWindow(state.npcID, msg, styles))
}

// SendShop packet to player
func (state *npcState) SendShop(goods [][]int32) {
	state.goods = goods
	state.conn.Send(packetShop(state.npcID, goods))
}

// Terminate the scriopt
func (state *npcState) Terminate() {
	state.terminate = true
}

// Selection value
func (state npcState) Selection() int32 {
	return state.selection
}

// InputString value
func (state npcState) InputString() string {
	return state.inputString
}

// InputNumber value
func (state npcState) InputNumber() int32 {
	return state.inputNumber
}

// Yes flag
func (state npcState) Yes() bool {
	return state.yes
}

// No flag
func (state npcState) No() bool {
	return state.no
}

// Next Flag
func (state npcState) Next() bool {
	return state.next
}

// Back flag
func (state npcState) Back() bool {
	return state.back
}

// Goods in the shop
func (state npcState) Goods() [][]int32 {
	return state.goods
}

// ClearFlags within the state
func (state *npcState) ClearFlags() {
	state.next = false
	state.back = false
	state.inputNumber = -1
	state.inputString = ""
	state.selection = -1
	state.yes = false
	state.no = false
}

// SetNextBack flags
func (state *npcState) SetNextBack(next, back bool) {
	state.next = next
	state.back = back
}

// SetYesNo flags
func (state *npcState) SetYesNo(yes, no bool) {
	state.yes = yes
	state.no = no
}

// SetTextInput option
func (state *npcState) SetTextInput(input string) {
	state.inputString = input
}

// SetNumberInput option
func (state *npcState) SetNumberInput(input int32) {
	state.inputNumber = input
}

// SetOptionSelect index
func (state *npcState) SetOptionSelect(selection int32) {
	state.selection = selection
}

// WarpPlayer to specific field
func (state npcState) WarpPlayer(p *player.Data, mapID int32) bool {
	if field, ok := state.fields[mapID]; ok {
		inst, err := field.GetInstance(0)

		if err != nil {
			return false
		}

		portal, err := inst.GetRandomSpawnPortal()

		state.warpFunc(p, field, portal)

		return true
	}

	return false
}

// GetInstance that the passed in player belongs to
func (state npcState) GetInstance(p *player.Data) *field.Instance {
	if field, ok := state.fields[p.MapID()]; ok {
		inst, err := field.GetInstance(p.InstanceID())

		if err != nil {
			return nil
		}

		return inst
	}

	return nil
}

// NpcChatController of the conversation
type NpcChatController struct {
	state npcState

	vm      *goja.Runtime
	program *goja.Program

	runFunc func(*npcState, *player.Data)
}

// CreateNewNpcController that will manage the npc conversation
func CreateNewNpcController(npcID int32, conn mnet.Client, program *goja.Program, warpFunc warpFn, fields map[int32]*field.Field) (*NpcChatController, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, err
	}

	controller := &NpcChatController{vm: vm, program: program}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", r)
		}
	}()

	err = vm.ExportTo(vm.Get("run"), &controller.runFunc)

	if err != nil {
		return nil, err
	}

	controller.state = npcState{npcID: npcID, conn: conn, warpFunc: warpFunc, fields: fields}

	return controller, nil
}

// Run the npc script
func (controller *NpcChatController) Run(p *player.Data) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in running script:", r)
		}
	}()

	controller.runFunc(&controller.state, p)

	return controller.state.terminate
}

// State struct of the npc
func (controller *NpcChatController) State() *npcState {
	return &controller.state
}
