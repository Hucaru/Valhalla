package npcchat

import (
	"fmt"
	"log"
	"strconv"

	"github.com/mattn/anko/packages"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/mattn/anko/core"

	"github.com/Hucaru/Valhalla/game/script"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/mattn/anko/vm"
)

type session struct {
	npcID  int32
	script string

	state       int
	isYes       bool
	selection   int
	stringInput string
	intInput    int
	style       int

	shopItemIndex   int16
	shopItemID      int32
	shopItemAmmount int16
	shopItems       [][]int32

	env *vm.Env
}

var sessions = make(map[mnet.MConnChannel]*session)

func NewSession(conn mnet.MConnChannel, npcID int32) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	contents, err := script.Get(strconv.Itoa(int(npcID)))

	if err != nil {
		contents =
			`if state == 1 {
				return SendOk("I have not been scripted. Please report #b` + strconv.Itoa(int(npcID)) + `#k on map #b` + strconv.Itoa(int(player.Char().MapID)) + `")
			}`

		fmt.Println(err)
	}

	sessions[conn] = &session{
		npcID:  npcID,
		script: contents,

		env: vm.NewEnv(),

		// NPC init state
		state:       1,
		isYes:       false,
		selection:   0,
		stringInput: "",
		intInput:    0,
	}

	packages.DefineImport(sessions[conn].env)

	sessions[conn].register(conn)
}

func NewSessionWithOverride(conn mnet.MConnChannel, script string, npcID int32) {
	sessions[conn] = &session{
		npcID:  npcID,
		script: script,

		env: vm.NewEnv(),

		// NPC init state
		state:       1,
		isYes:       false,
		selection:   0,
		stringInput: "",
		intInput:    0,
	}

	packages.DefineImport(sessions[conn].env)

	sessions[conn].register(conn)
}

func RemoveSession(conn mnet.MConnChannel) {
	delete(sessions, conn)
}

func Run(conn mnet.MConnChannel) {
	core.Import(sessions[conn].env)
	packet, err := sessions[conn].env.Execute(sessions[conn].script)

	if err != nil {
		log.Println(err)
	}

	p, ok := packet.(mpacket.Packet)

	if ok {
		conn.Send(p)
	} else {
		RemoveSession(conn)
	}
}

func Continue(conn mnet.MConnChannel, msgType, stateChange byte, reader mpacket.Reader) {
	if sessions[conn].state == 0 {
		RemoveSession(conn) // If we get here then remove the session
	} else {

		switch msgType {
		case 0:
			if stateChange == 1 {
				sessions[conn].state++
			} else if stateChange == 0xFF {
				sessions[conn].state = 0
			} else {
				sessions[conn].state--
			}

		case 1:
			sessions[conn].state += 1
			if stateChange == 0 {
				sessions[conn].isYes = false
			} else {
				sessions[conn].isYes = true
			}

		case 2:
			sessions[conn].state += 1

			if len(reader.GetRestAsBytes()) > 0 {
				sessions[conn].stringInput = string(reader.GetRestAsBytes())
			} else {
				sessions[conn].state = 0
			}

		case 3:
			sessions[conn].state += 1

			if len(reader.GetRestAsBytes()) > 0 {
				sessions[conn].intInput = int(reader.ReadUint32())
			} else {
				sessions[conn].state = 0
			}

		case 4:
			sessions[conn].state += 1

			if len(reader.GetRestAsBytes()) > 3 {
				sessions[conn].selection = int(reader.ReadUint32())
			} else {
				sessions[conn].state = 0
			}

		case 5:
			sessions[conn].state += 1

			// need to do
			fmt.Println("Finish this msg type: 5")

		default:
			log.Println("Unkown npc msg type:", msgType)
		}

		Run(conn)
	}
}

func Shop(conn mnet.MConnChannel, reader mpacket.Reader) {

}

func Storage(conn mnet.MConnChannel, reader mpacket.Reader) {

}
