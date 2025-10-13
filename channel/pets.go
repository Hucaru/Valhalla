package channel

import (
	"log"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type pet struct {
	name            string
	itemID          int32
	sn              int32
	itemDBID        int64
	level           byte
	closeness       int16
	fullness        byte
	deadDate        int64
	spawnDate       int64
	lastInteraction int64

	pos    pos
	stance byte

	spawned bool
}

func newPet(itemID, sn int32, dbID int64) *pet {
	itemInfo, err := nx.GetItem(itemID)
	if err != nil {
		log.Println(err)
	}

	return &pet{
		name:            itemInfo.Name,
		itemID:          itemID,
		sn:              sn,
		itemDBID:        dbID,
		stance:          0,
		level:           1,
		closeness:       0,
		fullness:        100,
		deadDate:        (time.Now().UnixMilli()*10000 + 116444592000000000 + (time.Hour.Milliseconds() * 24 * 90)),
		spawnDate:       0,
		lastInteraction: 0,
	}
}

func savePet(item *Item) error {
	// Initialize pet data if it doesn't exist
	if item.petData == nil {
		sn, _ := nx.GetCommoditySNByItemID(item.ID)
		item.petData = newPet(item.ID, sn, item.dbID)
	}

	p := item.petData

	_, err := common.DB.Exec(`
		INSERT INTO pets (
			parentID, name, sn, level, closeness, fullness,
			deadDate, spawnDate, lastInteraction
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			level = VALUES(level),
			closeness = VALUES(closeness),
			fullness = VALUES(fullness),
			deadDate = VALUES(deadDate),
			spawnDate = VALUES(spawnDate),
			lastInteraction = VALUES(lastInteraction)
	`, item.dbID,
		p.name,
		p.sn,
		p.level,
		p.closeness,
		p.fullness,
		p.deadDate,
		p.spawnDate,
		p.lastInteraction,
	)
	return err
}

func (p *pet) updateMovement(frag movementFrag) {
	p.pos.x = frag.x
	p.pos.y = frag.y
	p.pos.foothold = frag.foothold
	p.stance = frag.stance
}

func handlePetInteraction(plr *Player, pet *pet, interactionID byte, multiplier bool) bool {
	itm, err := nx.GetItem(pet.itemID)
	if err != nil || itm.Interact == nil {
		return false
	}
	react, ok := itm.Interact[interactionID]
	if !ok {
		return false
	}

	now := time.Now().UnixMilli()
	if now < pet.lastInteraction+15_000 || pet.level < react.LevelMin || pet.level > react.LevelMax || pet.fullness < 50 {
		return false
	}

	elapsed := float64(now - pet.lastInteraction - 15_000)
	pet.lastInteraction = now
	plr.MarkDirty(DirtyPet, time.Millisecond*300)

	mult := 1.0
	if multiplier && pet.name != "" {
		mult = 1.5
	}
	successProb := float64(react.Prob) * ((elapsed/10_000.0)*0.01 + 1) * mult
	success := float64(rand.Intn(100)) < successProb
	if success {
		pet.closeness += int16(react.Inc)
		if pet.closeness < 0 {
			pet.closeness = 0
		}
		if pet.closeness > 30_000 {
			pet.closeness = 30_000
		}
		pet.level = petLevelFromCloseness(pet.closeness)
		plr.updatePet()
	}
	return success
}

var thresholds = []int16{0, 1, 100, 300, 600, 1000, 1800, 3100, 5000, 8000, 12000, 17000, 22000, 28000}

func petLevelFromCloseness(c int16) byte {
	for lvl := byte(len(thresholds) - 1); lvl > 0; lvl-- {
		if c >= thresholds[lvl] {
			return lvl
		}
	}
	return 1
}

func packetPetAction(charID int32, op, action byte, text string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetAction)
	p.WriteInt32(charID)
	p.WriteByte(op)
	p.WriteByte(action)
	p.WriteString(text)
	return p
}

func packetPetNameChange(charID int32, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetNameChange)
	p.WriteInt32(charID)
	p.WriteString(name)
	return p
}

func packetPetInteraction(charID int32, interactionId byte, inc, food bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetInteraction)
	p.WriteInt32(charID)
	p.WriteBool(food)
	if !food {
		p.WriteByte(interactionId)
	}
	p.WriteBool(inc)

	return p
}

func packetPetMove(charID int32, move []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetMove)
	p.WriteInt32(charID)
	p.WriteBytes(move)
	return p
}

func packetPetSpawn(charID int32, petData *pet) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetSpawn)
	p.WriteInt32(charID)
	p.WriteBool(true)
	p.WriteInt32(petData.itemID)
	p.WriteString(petData.name)
	p.WriteUint64(uint64(petData.sn))
	p.WriteInt16(petData.pos.x)
	p.WriteInt16(petData.pos.y)
	p.WriteByte(petData.stance)
	p.WriteInt16(petData.pos.foothold)
	return p
}

func packetPetRemove(charID int32, reason byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPetSpawn)
	p.WriteInt32(charID)
	p.WriteBool(false)
	p.WriteByte(reason)

	return p
}
