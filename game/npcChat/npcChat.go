package npcChat

import (
	"github.com/Hucaru/Valhalla/game/script"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/mattn/anko/vm"
)

type session struct {
	npcID  string
	script string

	state       int
	isYes       bool
	selection   int
	stringInput string
	intInput    int

	env *vm.Env
}

var sessions map[mnet.MConnChannel]session

func NewSession(conn mnet.MConnChannel, npcID string) {
	contents, err := script.Get(npcID)

	if err != nil {
		contents =
			`if state == 1 {
				return SendOk('I have not been scripted. Please report #b" + strconv.Itoa(int(npcID)) + "#k on map #b" + strconv.Itoa(int(player.Char().mapID)) + "')
			}`
	}

	sessions[conn] = session{
		npcID:  scriptName,
		script: contents,

		state:       1,
		isYes:       false,
		selection:   0,
		stringInput: "",
		intInput:    0,

		env: vm.NewEnv(),
	}
}

func OverrideSessionScript(conn mnet.MConnChannel, script string) {
	if _, ok := sessions[conn]; ok {
		sessions[conn].scriptContent = script
	}
}

func RemoveSession(conn mnet.MConnChannel) {
	delete(sessions, conn)
}

func Run(conn mnet.MConnChannel) {

}
