package npcchat

import (
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (s *session) register() {
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
		return packet.NPCChatYesNo(s.npcID, msg)
	})

	s.env.Define("SendOk", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(s.npcID, msg, false, false)
	})

	s.env.Define("SendNext", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(s.npcID, msg, false, true)
	})

	s.env.Define("SendBackNext", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(s.npcID, msg, true, true)
	})

	s.env.Define("SendBack", func(msg string) mpacket.Packet {
		return packet.NPCChatBackNext(s.npcID, msg, true, false)
	})

	s.env.Define("SendUserStringInput", func(msg, defaultInput string, minLength, maxLength int16) mpacket.Packet {
		return packet.NPCChatUserString(s.npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendUserIntInput", func(msg string, defaultInput, minLength, maxLength int32) mpacket.Packet {
		return packet.NPCChatUserNumber(s.npcID, msg, defaultInput, minLength, maxLength)
	})

	s.env.Define("SendSelection", func(msg string) mpacket.Packet {
		return packet.NPCChatSelection(s.npcID, msg)
	})

	s.env.Define("SendStyleWindow", func(msg string, array []interface{}) mpacket.Packet {
		var styles []int32

		for _, i := range array {
			val, ok := i.(int64)

			if ok {
				styles = append(styles, int32(val))
			}
		}

		return packet.NPCChatStyleWindow(s.npcID, msg, styles)
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

		return packet.NPCShop(s.npcID, tmp)
	})
}
