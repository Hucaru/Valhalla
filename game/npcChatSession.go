package game

import (
	"fmt"
	"log"
	"strconv"

	"github.com/mattn/anko/packages"

	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/mattn/anko/core"

	"github.com/Hucaru/Valhalla/game/script"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/mattn/anko/vm"
)

type npcChatSession struct {
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

var npcChatSessions = make(map[mnet.Client]*npcChatSession)

func NewNpcChatSession(conn mnet.Client, npcID int32) {
	player, ok := Players[conn]

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

	npcChatSessions[conn] = &npcChatSession{
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

	packages.DefineImport(npcChatSessions[conn].env)

	npcChatSessions[conn].npcChatRegister(conn)
}

func NewNpcChatSessionWithOverride(conn mnet.Client, script string, npcID int32) {
	npcChatSessions[conn] = &npcChatSession{
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

	packages.DefineImport(npcChatSessions[conn].env)

	npcChatSessions[conn].npcChatRegister(conn)
}

func RemoveNpcChatSession(conn mnet.Client) {
	delete(npcChatSessions, conn)
}

func NpcChatRun(conn mnet.Client) {
	core.Import(npcChatSessions[conn].env)
	packet, err := npcChatSessions[conn].env.Execute(npcChatSessions[conn].script)

	if err != nil {
		log.Println(err)
	}

	p, ok := packet.(mpacket.Packet)

	if ok {
		conn.Send(p)
	} else {
		RemoveNpcChatSession(conn)
	}
}

func NpcChatContinue(conn mnet.Client, msgType, stateChange byte, reader mpacket.Reader) {
	if npcChatSessions[conn].state == 0 {
		RemoveNpcChatSession(conn) // If we get here then remove the session
	} else {
		switch msgType {
		case 0:
			if stateChange == 1 {
				npcChatSessions[conn].state++
			} else if stateChange == 0xFF {
				npcChatSessions[conn].state = 0
			} else {
				npcChatSessions[conn].state--
			}

		case 1:
			npcChatSessions[conn].state++
			if stateChange == 0 {
				npcChatSessions[conn].isYes = false
			} else {
				npcChatSessions[conn].isYes = true
			}

		case 2:
			npcChatSessions[conn].state++

			if len(reader.GetRestAsBytes()) > 0 {
				npcChatSessions[conn].stringInput = string(reader.GetRestAsBytes())
			} else {
				npcChatSessions[conn].state = 0
			}

		case 3:
			npcChatSessions[conn].state++

			if len(reader.GetRestAsBytes()) > 0 {
				npcChatSessions[conn].intInput = int(reader.ReadUint32())
			} else {
				npcChatSessions[conn].state = 0
			}

		case 4:
			npcChatSessions[conn].state++

			if len(reader.GetRestAsBytes()) > 3 {
				npcChatSessions[conn].selection = int(reader.ReadUint32())
			} else {
				npcChatSessions[conn].state = 0
			}

		case 5:
			npcChatSessions[conn].state++

			// need to do
			fmt.Println("Finish this msg type: 5")

		default:
			log.Println("Unkown npc msg type:", msgType)
		}

		NpcChatRun(conn)
	}
}

func NpcChatShop(conn mnet.Client, reader mpacket.Reader) {

}

func NpcChatStorage(conn mnet.Client, reader mpacket.Reader) {

}

func (s *npcChatSession) npcChatRegister(conn mnet.Client) {
	s.env.Define("state", &s.state)
	s.env.Define("isYes", &s.isYes)
	s.env.Define("selection", &s.selection)
	s.env.Define("stringInput", &s.stringInput)
	s.env.Define("intInput", &s.intInput)
	s.env.Define("style", &s.style)
	s.env.Define("shopItemIndex", &s.shopItemIndex)
	s.env.Define("shopItemID", &s.shopItemID)
	s.env.Define("shopItemAmmount", &s.shopItemAmmount)

	s.env.Define("SendYesNo", func(msg string) mpacket.Packet {
		return PacketNpcChatYesNo(s.npcID, msg)
	})

	s.env.Define("SendOk", func(msg string) mpacket.Packet {
		return PacketNpcChatBackNext(s.npcID, msg, false, false)
	})

	s.env.Define("SendNext", func(msg string) mpacket.Packet {
		return PacketNpcChatBackNext(s.npcID, msg, false, true)
	})

	s.env.Define("SendBackNext", func(msg string) mpacket.Packet {
		return PacketNpcChatBackNext(s.npcID, msg, true, true)
	})

	s.env.Define("SendBack", func(msg string) mpacket.Packet {
		return PacketNpcChatBackNext(s.npcID, msg, true, false)
	})

	s.env.Define("SendUserStringInput", func(msg, defaultInput string, minLength, maxLength int16) mpacket.Packet {
		return PacketNpcChatUserString(s.npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendUserIntInput", func(msg string, defaultInput, minLength, maxLength int32) mpacket.Packet {
		return PacketNpcChatUserNumber(s.npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendSelection", func(msg string) mpacket.Packet {
		return PacketNpcChatSelection(s.npcID, msg)
	})

	s.env.Define("SendStyleWindow", func(msg string, array []interface{}) mpacket.Packet {
		var styles []int32

		for _, i := range array {
			val, ok := i.(int64)

			if ok {
				styles = append(styles, int32(val))
			}
		}

		return PacketNpcChatStyleWindow(s.npcID, msg, styles)
	})

	s.env.Define("SendShop", func(items []interface{}) mpacket.Packet {
		tmp := make([][]int32, len(items))

		for i, v := range items {
			val := v.([]interface{})
			tmp[i] = []int32{}
			for _, j := range val {
				tmp[i] = append(tmp[i], int32(j.(int64)))
			}
		}
		s.shopItems = tmp

		return PacketNpcShop(s.npcID, tmp)
	})

	// Internal game logic
	s.env.Define("int32", func(i int64) int32 { return int32(i) })
	s.env.Define("player", Players[conn])
	s.env.Define("Maps", Maps)
	s.env.Define("Players", Players)
}
