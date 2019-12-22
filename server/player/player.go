package player

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Players type alias
type Players []Player

// GetFromConn retrieve the player from the connection
func (p Players) GetFromConn(conn mnet.Client) (*Player, error) {
	for _, v := range p {
		if v.conn == conn {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

// GetFromName retrieve the player from the connection
func (p Players) GetFromName(name string) (*Player, error) {
	for _, v := range p {
		if v.name == name {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

// GetFromID retrieve the player from the connection
func (p Players) GetFromID(id int32) (*Player, error) {
	for _, v := range p {
		if v.id == id {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

// RemoveFromConn removes the player based on the connection
func (p *Players) RemoveFromConn(conn mnet.Client) error {
	i := -1

	for j, v := range *p {
		if v.conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("Could not find player")
	}

	(*p)[i] = (*p)[len((*p))-1]
	(*p) = (*p)[:len((*p))-1]

	return nil
}

// Player connected to server
type Player struct {
	conn       mnet.Client
	instanceID int

	id        int32
	accountID int32
	worldID   byte

	mapID       int32
	mapPos      byte
	previousMap int32
	portalCount byte

	job int16

	level byte
	str   int16
	dex   int16
	intt  int16
	luk   int16
	hp    int16
	maxHP int16
	mp    int16
	maxMP int16
	ap    int16
	sp    int16
	exp   int32
	fame  int16

	name     string
	gender   byte
	skin     byte
	face     int32
	hair     int32
	chairID  int32
	stance   byte
	pos      pos
	foothold int16
	guild    string

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	inventory Inventory
	mesos     int32

	skills map[int32]Skill

	miniGameWins, miniGameDraw, miniGameLoss int32
}

func NewPlayer(conn mnet.Client, char Character) *Player {
	return &Player{conn: conn, char: char, instanceID: 0}
}

func (p Player) Conn() mnet.Client {
	return p.conn
}

func (p Player) Char() Character {
	return p
}

type Avatar interface {
	Name() string
	DisplayBytes() []byte
	MiniGameWins() int32
	MiniGameDraw() int32
	MiniGameLoss() int32
}

func (p Player) Avatar() Avatar {
	return p
}

func (p Player) InstanceID() int {
	return p.instanceID
}

func (p *Player) SetInstance(id int) {
	p.instanceID = id
}

func (p Player) Send(packet mpacket.Packet) {
	p.conn.Send(packet)
}

func (p *Player) SetJob(id int16) {
	p.job = id
	p.conn.Send(PacketPlayerStatChange(true, constant.JobID, int32(id)))
}

func (p *Player) levelUp(inst *instance) {
	p.GiveAP(5)
	p.GiveSP(3)

	levelUpHp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch p.job / 100 { // add effects from skills e.g. improve max mp
	case 0:
		p.maxHP += levelUpHp(constant.BeginnerHpAdd, 0)
		p.maxMP += levelUpMp(constant.BeginnerMpAdd, p.intt)
	case 1:
		p.maxHP += levelUpHp(constant.WarriorHpAdd, 0)
		p.maxMP += levelUpMp(constant.WarriorMpAdd, p.intt)
	case 2:
		p.maxHP += levelUpHp(constant.MagicianHpAdd, 0)
		p.maxMP += levelUpMp(constant.MagicianMpAdd, 2*p.intt)
	case 3:
		p.maxHP += levelUpHp(constant.BowmanHpAdd, 0)
		p.maxMP += levelUpMp(constant.BowmanMpAdd, p.intt)
	case 4:
		p.maxHP += levelUpHp(constant.ThiefHpAdd, 0)
		p.maxMP += levelUpMp(constant.ThiefMpAdd, p.intt)
	case 5:
		p.maxHP += constant.AdminHpAdd
		p.maxMP += constant.AdminMpAdd
	default:
		log.Println("Unkown job during level up", p.job)
	}

	p.hp = p.maxHP
	p.mp = p.maxMP

	p.SetHP(p.hp)
	p.SetMaxHP(p.hp)

	p.SetMP(p.mp)
	p.SetMaxMP(p.mp)

	p.GiveLevel(1, inst)
}

func (p *Player) SetEXP(amount int32, inst *instance) {
	if p.level > 199 {
		return
	}

	remainder := amount - constant.ExpTable[p.level-1]

	if remainder >= 0 {
		p.levelUp(inst)
		p.SetEXP(remainder, inst)
	} else {
		p.exp = amount
		p.Send(PacketPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

func (p *Player) GiveEXP(amount int32, fromMob, fromParty bool, inst *instance) {
	if fromMob {
		p.Send(PacketMessageExpGained(!fromParty, false, amount))
	} else {
		p.Send(PacketMessageExpGained(true, true, amount))
	}

	p.SetEXP(p.exp+amount, inst)
}

func (p *Player) SetLevel(amount byte, inst *instance) {
	p.level = amount
	p.Send(PacketPlayerStatChange(false, constant.LevelID, int32(amount)))
	inst.Send(PacketPlayerLevelUpAnimation(p.id))
}

func (p *Player) GiveLevel(amount byte, inst *instance) {
	p.SetLevel(p.level+amount, inst)
}

func (p *Player) SetAP(amount int16) {
	p.ap = amount
	p.Send(PacketPlayerStatChange(false, constant.ApID, int32(amount)))
}

func (p *Player) GiveAP(amount int16) {
	p.SetAP(p.ap + amount)
}

func (p *Player) SetSp(amount int16) {
	p.sp = amount
	p.Send(PacketPlayerStatChange(false, constant.SpID, int32(amount)))
}

func (p *Player) GiveSP(amount int16) {
	p.SetSp(p.sp + amount)
}

func (p *Player) SetStr(amount int16) {
	p.str = amount
	p.Send(PacketPlayerStatChange(true, constant.StrID, int32(amount)))
}

func (p *Player) GiveStr(amount int16) {
	p.SetStr(p.str + amount)
}

func (p *Player) SetDex(amount int16) {
	p.dex = amount
	p.Send(PacketPlayerStatChange(true, constant.DexID, int32(amount)))
}

func (p *Player) GiveDex(amount int16) {
	p.SetDex(p.dex + amount)
}

func (p *Player) SetInt(amount int16) {
	p.intt = amount
	p.Send(PacketPlayerStatChange(true, constant.IntID, int32(amount)))
}

func (p *Player) GiveInt(amount int16) {
	p.SetInt(p.intt + amount)
}

func (p *Player) SetLuk(amount int16) {
	p.luk = amount
	p.Send(PacketPlayerStatChange(true, constant.LukID, int32(amount)))
}

func (p *Player) GiveLuk(amount int16) {
	p.SetLuk(p.luk + amount)
}

func (p *Player) SetHP(amount int16) {
	p.hp = amount
	p.Send(PacketPlayerStatChange(true, constant.HpID, int32(amount)))
}

func (p *Player) GiveHP(amount int16) {
	p.SetHP(p.hp + amount)
	if p.hp < 0 {
		p.SetHP(0)
	}
}

func (p *Player) SetMaxHP(amount int16) {
	p.maxHP = amount
	p.Send(PacketPlayerStatChange(true, constant.MaxHpID, int32(amount)))
}

func (p *Player) SetMP(amount int16) {
	p.mp = amount
	p.Send(PacketPlayerStatChange(true, constant.MpID, int32(amount)))
}

func (p *Player) GiveMP(amount int16) {
	p.SetMP(p.mp + amount)
	if p.mp < 0 {
		p.SetMP(0)
	}
}

func (p *Player) SetMaxMP(amount int16) {
	p.maxMP = amount
	p.Send(PacketPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

func (p *Player) SetFame(amount int16) {

}

func (p *Player) SetGuild(name string) {

}

func (p *Player) SetEquipSlotSize(size byte) {

}

func (p *Player) SetUseSlotSize(size byte) {

}

func (p *Player) SetEtcSlotSize(size byte) {

}

func (p *Player) SetCashSlotSize(size byte) {

}

func (p *Player) SetMesos(amount int32) {
	p.mesos = amount
	p.Send(PacketPlayerStatChange(false, constant.MesosID, amount))
}

func (p *Player) GiveMesos(amount int32) {
	p.SetMesos(p.mesos + amount)
}

func (p *Player) SetMiniGameWins(v int32) {
	p.miniGameWins = v
}

func (p *Player) SetMiniGameLoss(v int32) {
	p.miniGameLoss = v
}

func (p *Player) SetMiniGameDraw(v int32) {
	p.miniGameDraw = v
}

func (p *Player) UpdateMovement(frag movementFrag) {
	p.pos.x = frag.x
	p.pos.y = frag.y
	p.foothold = frag.foothold
	p.stance = frag.stance
}

func (p *Player) SetPos(pos pos) {
	p.pos = pos
}

func (p Player) CheckPos(pos pos, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = p.pos.x == pos.x
	} else {
		xValid = (pos.x-xRange < p.pos.x && p.pos.x < pos.x+xRange)
	}

	if yRange == 0 {
		xValid = p.pos.y == pos.y
	} else {
		yValid = (pos.y-yRange < p.pos.y && p.pos.y < pos.y+yRange)
	}

	return xValid && yValid
}

func (p *Player) SetFoothold(fh int16) {
	p.foothold = fh
}

func (p *Player) SetMapID(id int32) {
	p.mapID = id
}

func (p *Player) SetMapPosID(pos byte) {
	p.mapPos = pos
}

func (p *Player) GiveItem(newItem item) error {
	findFirstEmptySlot := func(items []item, size byte) (int16, error) {
		slotsUsed := make([]bool, size)

		for _, v := range items {
			if v.slotID > 0 {
				slotsUsed[v.slotID-1] = true
			}
		}

		slot := 0

		for i, v := range slotsUsed {
			if v == false {
				slot = i + 1
				break
			}
		}

		if slot == 0 {
			slot = len(slotsUsed) + 1
		}

		if byte(slot) > size {
			return 0, fmt.Errorf("No empty item slot left")
		}

		return int16(slot), nil
	}

	switch newItem.invID {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(p.inventory.equip, p.equipSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.amount = 1 // just in case
		p.inventory.equip = append(p.inventory.equip, newItem)
		p.Send(PacketInventoryAddItem(newItem, true))
	case 2: // Use
		var slotID int16
		var index int
		for i, v := range p.inventory.use {
			if v.itemID == newItem.itemID && v.amount < constant.MaxItemStack {
				slotID = v.slotID
				index = i
				break
			}
		}

		if slotID == 0 {
			slotID, err := findFirstEmptySlot(p.inventory.use, p.useSlotSize)

			if err != nil {
				return err
			}

			newItem.slotID = slotID
			p.inventory.use = append(p.inventory.use, newItem)
			p.Send(PacketInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.amount - (constant.MaxItemStack - p.inventory.use[index].amount)

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(p.inventory.use, p.useSlotSize)

				if err != nil {
					return err
				}

				newItem.amount = remainder
				newItem.slotID = slotID
				p.inventory.use = append(p.inventory.use, newItem)
				p.inventory.use[index].amount = constant.MaxItemStack

				p.Send(PacketInventoryAddItems([]item{p.inventory.use[index], newItem}, []bool{false, true}))
			} else { // full merge
				p.inventory.use[index].amount += newItem.amount
				p.Send(PacketInventoryAddItem(p.inventory.use[index], false))
			}
		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(p.inventory.setUp, p.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		p.inventory.setUp = append(p.inventory.setUp, newItem)
		p.Send(PacketInventoryAddItem(newItem, true))
	case 4: // Etc
		var slotID int16
		var index int
		for i, v := range p.inventory.etc {
			if v.itemID == newItem.itemID && v.amount < constant.MaxItemStack {
				slotID = v.slotID
				index = i
				break
			}
		}

		if slotID == 0 {
			slotID, err := findFirstEmptySlot(p.inventory.etc, p.etcSlotSize)

			if err != nil {
				return err
			}

			newItem.slotID = slotID
			p.inventory.etc = append(p.inventory.etc, newItem)
			p.Send(PacketInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.amount - (constant.MaxItemStack - p.inventory.etc[index].amount)

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(p.inventory.etc, p.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.amount = remainder
				newItem.slotID = slotID
				p.inventory.etc = append(p.inventory.etc, newItem)
				p.inventory.etc[index].amount = constant.MaxItemStack

				p.Send(PacketInventoryAddItems([]item{p.inventory.etc[index], newItem}, []bool{false, true}))
			} else { // full merge
				p.inventory.etc[index].amount += newItem.amount
				p.Send(PacketInventoryAddItem(p.inventory.etc[index], false))
			}
		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(p.inventory.cash, p.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		p.inventory.cash = append(p.inventory.cash, newItem)
		p.Send(PacketInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.invID)
	}
	return nil
}

func (p *Player) TakeItem(itemID int32, amount int16) (item, error) {
	return item{}, nil
}

func (p *Player) RemoveItem(remove item) {
	findIndex := func(items []item, item item) int {
		for i, v := range items {
			if v.uuid == remove.uuid {
				return i
			}
		}

		return 0
	}

	switch remove.invID {
	case 1:
		if i := findIndex(p.inventory.equip, remove); i != 0 {
			p.inventory.equip[i] = p.inventory.equip[len(p.inventory.equip)-1]
			p.inventory.equip = p.inventory.equip[:len(p.inventory.equip)-1]
		}
	case 2:
		if i := findIndex(p.inventory.use, remove); i != 0 {
			p.inventory.use[i] = p.inventory.use[len(p.inventory.use)-1]
			p.inventory.use = p.inventory.use[:len(p.inventory.use)-1]
		}
	case 3:
		if i := findIndex(p.inventory.setUp, remove); i != 0 {
			p.inventory.setUp[i] = p.inventory.setUp[len(p.inventory.setUp)-1]
			p.inventory.setUp = p.inventory.setUp[:len(p.inventory.setUp)-1]
		}
	case 4:
		if i := findIndex(p.inventory.etc, remove); i != 0 {
			p.inventory.etc[i] = p.inventory.etc[len(p.inventory.etc)-1]
			p.inventory.etc = p.inventory.etc[:len(p.inventory.etc)-1]
		}
	case 5:
		if i := findIndex(p.inventory.cash, remove); i != 0 {
			p.inventory.cash[i] = p.inventory.cash[len(p.inventory.cash)-1]
			p.inventory.cash = p.inventory.cash[:len(p.inventory.cash)-1]
		}
	}
}

func (p Player) GetItem(invID byte, slotID int16) (item, error) {
	var result item
	var err error

	findItem := func(items []item, slotID int16) (item, error) {
		for _, v := range items {
			if v.slotID == slotID {
				return v, nil
			}
		}

		return item{}, fmt.Errorf("Unable to get item")
	}

	switch invID {
	case 1:
		result, err = findItem(p.inventory.equip, slotID)
	case 2:
		result, err = findItem(p.inventory.use, slotID)
	case 3:
		result, err = findItem(p.inventory.setUp, slotID)
	case 4:
		result, err = findItem(p.inventory.etc, slotID)
	case 5:
		result, err = findItem(p.inventory.cash, slotID)
	}

	return result, err
}

func (p *Player) UpdateItem(orig, new item) {
	var items []item

	switch new.invID {
	case 1:
		items = p.inventory.equip
	case 2:
		items = p.inventory.use
	case 3:
		items = p.inventory.setUp
	case 4:
		items = p.inventory.etc
	case 5:
		items = p.inventory.cash
	}

	for i, v := range items {
		if v.uuid == new.uuid {
			items[i] = new
			break
		}
	}
}

func (p *Player) UpdateSkill(updatedSkill Skill) {
	p.skills[updatedSkill.ID] = updatedSkill
	p.Send(PacketPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
}

func (c Player) ID() int32               { return p.id }
func (p Player) AccountID() int32        { return p.accountID }
func (p Player) WorldID() byte           { return p.worldID }
func (p Player) MapID() int32            { return p.mapID }
func (p Player) MapPos() byte            { return p.mapPos }
func (p Player) PreviousMap() int32      { return p.previousMap }
func (p Player) PortalCount() byte       { return p.portalCount }
func (p Player) Job() int16              { return p.job }
func (p Player) Level() byte             { return p.level }
func (p Player) Str() int16              { return p.str }
func (p Player) Dex() int16              { return p.dex }
func (p Player) Int() int16              { return p.intt }
func (p Player) Luk() int16              { return p.luk }
func (p Player) HP() int16               { return p.hp }
func (p Player) MaxHP() int16            { return p.maxHP }
func (p Player) MP() int16               { return p.mp }
func (p Player) MaxMP() int16            { return p.maxMP }
func (p Player) AP() int16               { return p.ap }
func (p Player) SP() int16               { return p.sp }
func (p Player) Exp() int32              { return p.exp }
func (p Player) Fame() int16             { return p.fame }
func (p Player) Name() string            { return p.name }
func (p Player) Gender() byte            { return p.gender }
func (p Player) Skin() byte              { return p.skin }
func (p Player) Face() int32             { return p.face }
func (p Player) Hair() int32             { return p.hair }
func (p Player) ChairID() int32          { return p.chairID }
func (p Player) Stance() byte            { return p.stance }
func (p Player) Pos() pos                { return p.pos }
func (p Player) Foothold() int16         { return p.foothold }
func (p Player) Guild() string           { return p.guild }
func (p Player) EquipSlotSize() byte     { return p.equipSlotSize }
func (p Player) UseSlotSize() byte       { return p.useSlotSize }
func (p Player) SetupSlotSize() byte     { return p.setupSlotSize }
func (p Player) EtcSlotSize() byte       { return p.etcSlotSize }
func (p Player) CashSlotSize() byte      { return p.cashSlotSize }
func (p Player) Inventory() Inventory    { return p.inventory }
func (p Player) Mesos() int32            { return p.mesos }
func (p Player) Skills() map[int32]Skill { return p.skills }
func (p Player) MiniGameWins() int32     { return p.miniGameWins }
func (p Player) MiniGameDraw() int32     { return p.miniGameDraw }
func (p Player) MiniGameLoss() int32     { return p.miniGameLoss }
func (plr Player) DisplayBytes() []byte {
	p := mpacket.NewPacket()
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteByte(0x00) // ?
	p.WriteInt32(plr.hair)

	cashWeapon := int32(0)

	for _, b := range plr.inventory.equip {
		if b.slotID < 0 && b.slotID > -20 {
			p.WriteByte(byte(math.Abs(float64(b.slotID))))
			p.WriteInt32(b.itemID)
		}
	}

	for _, b := range plr.inventory.equip {
		if b.slotID < -100 {
			if b.slotID == -111 {
				cashWeapon = b.itemID
			} else {
				p.WriteByte(byte(math.Abs(float64(b.slotID + 100))))
				p.WriteInt32(b.itemID)
			}
		}
	}

	p.WriteByte(0xFF)
	p.WriteByte(0xFF)
	p.WriteInt32(cashWeapon)

	return p
}
