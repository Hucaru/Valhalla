package script

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/server/player"
	"github.com/dop251/goja"
)

type npcState struct {
	npcID       int32
	conn        mnet.Client
	terminate   bool
	selection   int32
	inputString string
	inputNumber int32

	// flags
	yes, no, next, back bool

	goods [][]int32
}

func (state *npcState) SendBackNext(msg string, back, next bool) {
	state.conn.Send(packetChatBackNext(state.npcID, msg, next, back))
}

func (state *npcState) SendOK(msg string) {
	state.conn.Send(packetChatOk(state.npcID, msg))
}

func (state *npcState) SendYesNo(msg string) {
	state.conn.Send(packetChatYesNo(state.npcID, msg))
}

func (state *npcState) SendInputText(msg, defaultInput string, minLength, maxLength int16) {
	state.conn.Send(packetChatUserString(state.npcID, msg, defaultInput, minLength, maxLength))
}

func (state *npcState) SendInputNumber(msg string, defaultInput, minLength, maxLength int32) {
	state.conn.Send(packetChatUserNumber(state.npcID, msg, defaultInput, minLength, maxLength))
}

func (state *npcState) SendSelection(msg string) {
	state.conn.Send(packetChatSelection(state.npcID, msg))
}

func (state *npcState) SendStyles(msg string, styles []int32) {
	state.conn.Send(packetChatStyleWindow(state.npcID, msg, styles))
}

func (state *npcState) SendShop(goods [][]int32) {
	state.goods = goods
	state.conn.Send(packetShop(state.npcID, goods))
}

func (state *npcState) Terminate() {
	state.terminate = true
}

func (state npcState) Selection() int32 {
	return state.selection
}

func (state npcState) InputString() string {
	return state.inputString
}

func (state npcState) InputNumber() int32 {
	return state.inputNumber
}

func (state npcState) Yes() bool {
	return state.yes
}

func (state npcState) No() bool {
	return state.no
}

func (state npcState) Next() bool {
	return state.next
}

func (state npcState) Back() bool {
	return state.back
}

func (state npcState) Goods() [][]int32 {
	return state.goods
}

// NpcChatController of the conversation
type NpcChatController struct {
	state npcState

	vm      *goja.Runtime
	program *goja.Program

	runFunc func(*npcState, *player.Data)
}

// CreateNewNpcController that will manage the npc conversation
func CreateNewNpcController(npcID int32, conn mnet.Client, program *goja.Program) (*NpcChatController, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	_, err := vm.RunProgram(program)

	if err != nil {
		return nil, err
	}

	controller := &NpcChatController{vm: vm, program: program}

	err = vm.ExportTo(vm.Get("run"), &controller.runFunc)

	if err != nil {
		return nil, err
	}

	controller.state = npcState{npcID: npcID, conn: conn}

	return controller, nil
}

// Run the npc script
func (controller *NpcChatController) Run(p *player.Data) bool {
	controller.runFunc(&controller.state, p)

	return controller.state.terminate
}

// State struct of the npc
func (controller *NpcChatController) State() npcState {
	return controller.state
}

// ClearFlags within the state
func (controller *NpcChatController) ClearFlags() {
	controller.state.next = false
	controller.state.back = false
	controller.state.inputNumber = -1
	controller.state.inputString = ""
	controller.state.selection = -1
	controller.state.yes = false
	controller.state.no = false
}

// SetNextBack flags
func (controller *NpcChatController) SetNextBack(next, back bool) {
	controller.state.next = next
	controller.state.back = back
}

// SetYesNo flags
func (controller *NpcChatController) SetYesNo(yes, no bool) {
	controller.state.yes = yes
	controller.state.no = no
}

// SetTextInput option
func (controller *NpcChatController) SetTextInput(input string) {
	controller.state.inputString = input
}

// SetNumberInput option
func (controller *NpcChatController) SetNumberInput(input int32) {
	controller.state.inputNumber = input
}

// SetOptionSelect index
func (controller *NpcChatController) SetOptionSelect(selection int32) {
	controller.state.selection = selection
}
