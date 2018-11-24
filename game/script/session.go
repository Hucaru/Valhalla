package npcdialogue

import (
	"log"
	"strconv"

	"github.com/Hucaru/Valhalla/game"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/vm"
)

var sessions = make(map[mnet.MConnChannel]*session)
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

func NewSession(conn mnet.MConnChannel, npcID int32, player game.Player) {
	var script string

	if _, exists := scripts[npcID]; exists {
		script = scripts[npcID]
	} else {
		script = "if state == 1 {return SendOk('I have not been scripted. Please report #b" + strconv.Itoa(int(npcID)) + "#k on map #b" + strconv.Itoa(int(char.GetCurrentMap())) + "')}"
	}

	sessions[conn] = &session{conn: conn,
		state:       1,
		isYes:       false,
		selection:   0,
		stringInput: "",
		intInput:    0,
		script:      script,
		env:         vm.NewEnv(),
		npcID:       npcID}

	sessions[conn].register(npcID, char)
}

func (s *session) OverrideScript(script string) {
	s.script = script
}

func RemoveSession(conn mnet.MConnChannel) {
	delete(sessions, conn)
}

func GetSession(conn mnet.MConnChannel) *session {
	result := sessions[conn]

	return result
}

type session struct {
	conn mnet.MConnChannel

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
	s.env.Define("SendYesNo", func(msg string) mpacket.Packet {
		return packet.NPCChatYesNo(npcID, msg)
	})

	s.env.Define("SendOk", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(npcID, msg, false, false)
	})

	s.env.Define("SendNext", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(npcID, msg, false, true)
	})

	s.env.Define("SendBackNext", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(npcID, msg, true, true)
	})

	s.env.Define("SendBack", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(npcID, msg, true, false)
	})

	s.env.Define("SendUserStringInput", func(msg, defaultInput string, minLength, maxLength int16) mpacket.Packet {
		return packet.NPCChatUserString(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendUserIntInput", func(msg string, defaultInput, minLength, maxLength int32) mpacket.Packet {
		return packet.NPCChatUserNumber(npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendSelection", func(msg string) mpacket.Packet {
		return packet.NPCChatSelection(npcID, msg)
	})

	s.env.Define("SendStyleWindow", func(msg string, array []interface{}) mpacket.Packet {
		var styles []int32

		for _, i := range array {
			val, ok := i.(int64)

			if ok {
				styles = append(styles, int32(val))
			}
		}

		return packet.NPCChatStyleWindow(npcID, msg, styles)
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

		return packet.NPCShop(npcID, tmp)
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

	p, ok := packet.(mpacket.Packet)

	if ok {
		s.conn.Send(p)
	} else {
		s.state = 1 // should probably delete the session here
	}
}

func (s *session) Continue(msgType byte, stateChange byte, reader mpacket.Reader) {

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

func (s *session) Shop(reader mpacket.Reader) {
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
						s.conn.Send(packet.NPCShopNotEnoughMesos())
					} else {
						newItem, _ := inventory.CreateFromID(info[0], false)
						newItem.Amount = s.shopItemAmmount

						if char.GiveItem(newItem) {
							char.TakeMesos(price)
							s.conn.Send(packet.NPCShopContinue())
						} else {
							s.conn.Send(packet.NPCShopNotEnoughStock())
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
				if item.ItemID == itemID && item.SlotID == slotID {
					if ammount > item.Amount {
						break
					}

					if char.TakeItem(item, ammount) {
						char.GiveMesos(nx.Items[itemID].Price * int32(ammount))
					}

					s.conn.Send(packet.NPCShopContinue())
					break
				}
			}
		})
	case 2: // recharge
		slotID := reader.ReadInt16()

		channel.Players.OnCharacterFromConn(s.conn, func(char *channel.MapleCharacter) {
			for _, curItem := range char.GetItems() {
				if curItem.InvID == 2 && curItem.SlotID == slotID && inventory.IsRechargeAble(curItem.ItemID) {
					price := int32(nx.Items[curItem.ItemID].UnitPrice * float64(nx.Items[curItem.ItemID].SlotMax))

					if price > char.GetMesos() {
						s.conn.Send(packet.NPCShopNotEnoughMesos())
					} else {
						curItem.Amount = int16(nx.Items[curItem.ItemID].SlotMax)
						char.UpdateItem(curItem)
						char.TakeMesos(price)
						s.conn.Send(packet.NPCShopContinue())
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

func (s *session) Storage(reader mpacket.Reader) {

}
