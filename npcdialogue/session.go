package npcdialogue

import (
	"log"
	"strconv"
	"sync"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/vm"
)

var sessionsMutex = &sync.RWMutex{}
var scriptsMutex = &sync.RWMutex{}

var sessions = make(map[interop.ClientConn]*session)
var scripts = make(map[int32]string)

func addScript(npcID int32, contents string) {
	scripts[npcID] = contents
}

func removeScripts(npcID int32) {
	delete(scripts, npcID)
}

func init() {
	// Load scripts
	loadScripts()
	go watchFiles()
}

func NewSession(conn interop.ClientConn, npcID int32, char *channel.MapleCharacter) {
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

	shopItemIndex   int16
	shopItemID      int32
	shopItemAmmount int16

	shopItems [][]int32

	script string
	env    *vm.Env
	npcID  int32
}

func (s *session) register(npcID int32, char *channel.MapleCharacter) {
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

	s.env.Define("SendUserStringInput", func(msg, defaultInput string, minLength, maxLength int16) maplepacket.Packet {
		return packets.NPCChatUserString(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendUserIntInput", func(msg string, defaultInput, minLength, maxLength int32) maplepacket.Packet {
		return packets.NPCChatUserNumber(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendSelection", func(msg string) maplepacket.Packet {
		return packets.NPCChatSelection(npcID, msg)
	})

	s.env.Define("SendStyleWindow", func(msg string, array []interface{}) maplepacket.Packet {
		var styles []int32

		for _, i := range array {
			val, ok := i.(int64)

			if ok {
				styles = append(styles, int32(val))
			}
		}

		return packets.NPCChatStyleWindow(npcID, msg, styles)
	})

	s.env.Define("SendShop", func(items []interface{}) maplepacket.Packet {
		tmp := make([][]int32, len(items))

		for i, v := range items {
			val := v.([]interface{})
			tmp[i] = []int32{}
			for _, j := range val {
				tmp[i] = append(tmp[i], int32(j.(int64)))
			}
		}
		s.shopItems = tmp

		return packets.NPCShop(npcID, tmp)
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
		s.state = 1 // should probably delete the session here
	}
}

func (s *session) Continue(msgType byte, stateChange byte, reader maplepacket.Reader) {

	if s.state == 0 {

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

			// need to do

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
	case 0: // Buy item
		s.shopItemIndex = reader.ReadInt16()
		s.shopItemID = reader.ReadInt32()
		s.shopItemAmmount = reader.ReadInt16()

		shopInd := int16(-1)

		for _, info := range s.shopItems {
			if len(info) == 2 && info[1] == 0 { // Rechargeables do not count as part of the ind counter
				continue
			}

			shopInd++

			if shopInd == s.shopItemIndex && info[0] == s.shopItemID {
				channel.Players.OnCharacterFromConn(s.conn, func(char *channel.MapleCharacter) {
					price := nx.Items[info[0]].Price

					if len(info) == 2 {
						price = info[1] * int32(s.shopItemAmmount)
					}

					if price > char.GetMesos() {
						// client is supposed to protect against this
					} else {
						newItem := inventory.CreateFromID(info[0], false)
						newItem.SetAmount(s.shopItemAmmount)

						if char.GiveItem(newItem) {
							char.TakeMesos(price)
						} else {
							s.conn.Write(packets.NPCShopNotEnoughStock())
						}
					}
				})
			}
		}
	case 1: // sell item
		slotID := reader.ReadInt16()
		itemID := reader.ReadInt32()
		ammount := reader.ReadInt16()

		channel.Players.OnCharacterFromConn(s.conn, func(char *channel.MapleCharacter) {
			for _, item := range char.GetItems() {
				if item.GetItemID() == itemID && item.GetSlotID() == slotID {
					if ammount > item.GetAmount() {
						break
					}

					char.TakeItem(item.GetInvID(), slotID, ammount)
					char.GiveMesos(nx.Items[itemID].Price * int32(ammount))
					break
				}
			}
		})
	case 2: // recharge
		slotID := reader.ReadInt16()

		channel.Players.OnCharacterFromConn(s.conn, func(char *channel.MapleCharacter) {
			for _, currentItem := range char.GetItems() {
				if currentItem.GetInvID() == 2 && currentItem.GetSlotID() == slotID && inventory.IsRechargeAble(currentItem.GetItemID()) {
					price := int32(nx.Items[currentItem.GetItemID()].UnitPrice * float64(nx.Items[currentItem.GetItemID()].SlotMax))

					if price > char.GetMesos() {
						s.conn.Write(packets.NPCShopNotEnoughMesos())
					} else {
						currentItem.SetAmount(int16(nx.Items[currentItem.GetItemID()].SlotMax))
						char.GiveItem(currentItem)
						char.TakeMesos(price)
					}
				}
			}
		})
	case 3:
		// closed window, nothing to handle here, state system takes care of it
	default:
		log.Println("Unkown shop operation:", operation, reader)
	}

	s.Run()
}

func (s *session) Storage(reader maplepacket.Reader) {

}
