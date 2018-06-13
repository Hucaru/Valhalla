package npcChat

import (
	"log"
	"strconv"
	"sync"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/vm"
)

var sessionsMutex = &sync.RWMutex{}
var scriptsMutex = &sync.RWMutex{}

var sessions = make(map[interop.ClientConn]*session)
var scripts = make(map[uint32]string)

func addScript(npcID uint32, contents string) {
	scripts[npcID] = contents
}

func removeScripts(npcID uint32) {
	delete(scripts, npcID)
}

func init() {
	// Load scripts
	loadScripts()
	go watchFiles()
}

func NewSession(conn interop.ClientConn, npcID uint32, char *channel.MapleCharacter) {
	var script string

	scriptsMutex.RLock()
	if _, exists := scripts[npcID]; exists {
		script = scripts[npcID]
	} else {
		script = "if state == 1 {return SendOk('I have not been scripted. Please report #b" + strconv.Itoa(int(npcID)) + "#k on map #b" + strconv.Itoa(int(char.GetCurrentMap())) + "')}"
	}
	scriptsMutex.RUnlock()

	sessionsMutex.Lock()

	sessions[conn] = &session{conn: conn,
		state:       1,
		isYes:       false,
		selection:   0,
		stringInput: "",
		intInput:    0,
		script:      script,
		env:         vm.NewEnv(),
		npcID:       npcID}
	sessionsMutex.Unlock()

	scriptsMutex.RLock()
	sessions[conn].register(npcID, char)
	scriptsMutex.RUnlock()
}

func RemoveSession(conn interop.ClientConn) {
	sessionsMutex.Lock()
	delete(sessions, conn)
	sessionsMutex.Unlock()
}

func GetSession(conn interop.ClientConn) *session {
	scriptsMutex.RLock()
	result := sessions[conn]
	scriptsMutex.RUnlock()

	return result
}

type session struct {
	conn interop.ClientConn

	state       int
	isYes       bool
	selection   int
	stringInput string
	intInput    int
	style       int

	shopItemIndex   uint16
	shopItemID      uint32
	shopItemAmmount uint16

	script string
	env    *vm.Env
	npcID  uint32
}

func (s *session) register(npcID uint32, char *channel.MapleCharacter) {
	s.env.Define("SendYesNo", func(msg string) maplepacket.Packet {
		return packets.NPCChatYesNo(npcID, msg)
	})

	s.env.Define("SendOk", func(msg string) maplepacket.Packet {
		return packets.NPCChatBackNext(npcID, msg, false, false)
	})

	s.env.Define("SendNext", func(msg string) maplepacket.Packet {
		return packets.NPCChatBackNext(npcID, msg, false, true)
	})

	s.env.Define("SendBackNext", func(msg string) maplepacket.Packet {
		return packets.NPCChatBackNext(npcID, msg, true, true)
	})

	s.env.Define("SendBack", func(msg string) maplepacket.Packet {
		return packets.NPCChatBackNext(npcID, msg, true, false)
	})

	s.env.Define("SendUserStringInput", func(msg, defaultInput string, minLength, maxLength uint16) maplepacket.Packet {
		return packets.NPCChatUserString(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendUserIntInput", func(msg string, defaultInput, minLength, maxLength uint32) maplepacket.Packet {
		return packets.NPCChatUserNumber(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendSelection", func(msg string) maplepacket.Packet {
		return packets.NPCChatSelection(npcID, msg)
	})

	s.env.Define("SendStyleWindow", func(msg string, array []interface{}) maplepacket.Packet {
		var styles []uint32

		for _, i := range array {
			val, ok := i.(int64)

			if ok {
				styles = append(styles, uint32(val))
			}
		}

		return packets.NPCChatStyleWindow(npcID, msg, styles)
	})

	s.env.Define("SendShop", func(items map[uint32]uint32) maplepacket.Packet {
		return packets.NPCShop(npcID, items)
	})

	s.env.Define("SendPacketToMap", channel.Maps.GetMap(char.GetCurrentMap()).SendPacket)

	s.env.Define("state", &s.state)
	s.env.Define("isYes", &s.isYes)
	s.env.Define("selection", &s.selection)
	s.env.Define("stringInput", &s.stringInput)
	s.env.Define("intInput", &s.intInput)
	s.env.Define("style", &s.style)
	s.env.Define("shopItemIndex", &s.shopItemIndex)
	s.env.Define("shopItemID", &s.shopItemID)
	s.env.Define("shopItemAmmount", &s.shopItemAmmount)

	s.env.Define("player", char)
	s.env.Define("maps", &channel.Maps)

}

func (s *session) Run() {
	core.Import(s.env)
	packet, err := s.env.Execute(s.script)

	if err != nil {
		log.Println(err)
	}

	p, ok := packet.(maplepacket.Packet)

	if ok {
		s.conn.Write(p)
	} else {
		s.state = 1
	}
}

func (s *session) Continue(msgType byte, stateChange byte, reader maplepacket.Reader) {

	if s.state == 0 {
		sessionsMutex.Lock()
		delete(sessions, s.conn)
		sessionsMutex.Unlock()
	} else {

		switch msgType {
		case 0:
			if stateChange == 1 {
				s.state += 1
			} else if stateChange == 0xFF {
				s.state = 0
			} else {
				s.state -= 1
			}

		case 1:
			s.state += 1
			if stateChange == 0 {
				s.isYes = false
			} else {
				s.isYes = true
			}

		case 2:
			s.state += 1

			if len(reader.GetRestAsBytes()) > 0 {
				s.stringInput = string(reader.GetRestAsBytes())
			} else {
				s.state = 0
			}

		case 3:
			s.state += 1

			if len(reader.GetRestAsBytes()) > 0 {
				s.intInput = int(reader.ReadUint32())
			} else {
				s.state = 0
			}

		case 4:
			s.state += 1

			if len(reader.GetRestAsBytes()) > 3 {
				s.selection = int(reader.ReadUint32())
			} else {
				s.state = 0
			}

		case 5:
			s.state += 1

			// need to

		default:
			log.Println("Unkown npc msg type:", msgType)
		}

		s.Run()
	}
}

func (s *session) Shop(reader maplepacket.Reader) {
	operation := reader.ReadByte() // ?

	s.state += 1

	switch operation {
	case 0:
		s.shopItemIndex = reader.ReadUint16()
		s.shopItemID = reader.ReadUint32()
		s.shopItemAmmount = reader.ReadUint16()
	case 3:
	default:
		log.Println("Unkown shop operation:", operation)
	}
	reader.ReadUint16()
}

func (s *session) Storage(reader maplepacket.Reader) {

}
