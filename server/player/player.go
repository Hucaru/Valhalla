package player

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// TODO: Move Players into server level logic

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

type item interface {
	ID() int32
	DbID() int64
	Save(*sql.DB, int32)
	Cash() bool
	InvID() byte
	SlotID() int16
	SetSlotID(int16)
	Amount() int16
	SetAmount(int16)
	InventoryBytes() []byte
	ShortBytes() []byte
}

type pos interface {
	X() int16
	SetX(int16)
	Y() int16
	SetY(int16)
}

type portal interface {
	ID() byte
}

type instance interface {
	Send(mpacket.Packet)
	CalculateNearestSpawnPortal(pos) (portal, error)
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

	equip []item
	use   []item
	setUp []item
	etc   []item
	cash  []item

	mesos int32

	skills map[int32]Skill

	miniGameWins, miniGameDraw, miniGameLoss int32
}

// NewPlayer - returns a player struct from a client connection and inventory
func NewPlayer(conn mnet.Client, equip []item, use []item, setUp []item, etc []item, cash []item) Player {
	return Player{conn: conn, instanceID: 0, equip: equip, use: use, setUp: setUp, etc: etc, cash: cash}
}

// Conn - client connection associated with this player
func (p Player) Conn() mnet.Client {
	return p.conn
}

// InstanceID - field instance id the player is currently on
func (p Player) InstanceID() int {
	return p.instanceID
}

// SetInstance - assign the instance id for the player
func (p *Player) SetInstance(id int) {
	p.instanceID = id
}

// Send the player a packet
func (p Player) Send(packet mpacket.Packet) {
	p.conn.Send(packet)
}

// SetJob id of the player
func (p *Player) SetJob(id int16) {
	p.job = id
	p.conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))
}

func (p *Player) levelUp(inst instance) {
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

// SetEXP of the player
func (p *Player) SetEXP(amount int32, inst instance) {
	if p.level > 199 {
		return
	}

	remainder := amount - constant.ExpTable[p.level-1]

	if remainder >= 0 {
		p.levelUp(inst)
		p.SetEXP(remainder, inst)
	} else {
		p.exp = amount
		p.Send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

// GiveEXP to the player
func (p *Player) GiveEXP(amount int32, fromMob, fromParty bool, inst instance) {
	if fromMob {
		p.Send(packetMessageExpGained(!fromParty, false, amount))
	} else {
		p.Send(packetMessageExpGained(true, true, amount))
	}

	p.SetEXP(p.exp+amount, inst)
}

// SetLevel of the player
func (p *Player) SetLevel(amount byte, inst instance) {
	p.level = amount
	p.Send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	inst.Send(packetPlayerLevelUpAnimation(p.id))
}

// GiveLevel amount ot the player
func (p *Player) GiveLevel(amount byte, inst instance) {
	p.SetLevel(p.level+amount, inst)
}

// SetAP of player
func (p *Player) SetAP(amount int16) {
	p.ap = amount
	p.Send(packetPlayerStatChange(false, constant.ApID, int32(amount)))
}

// GiveAP to player
func (p *Player) GiveAP(amount int16) {
	p.SetAP(p.ap + amount)
}

// SetSP of player
func (p *Player) SetSP(amount int16) {
	p.sp = amount
	p.Send(packetPlayerStatChange(false, constant.SpID, int32(amount)))
}

// GiveSP to player
func (p *Player) GiveSP(amount int16) {
	p.SetSP(p.sp + amount)
}

// SetStr of the player
func (p *Player) SetStr(amount int16) {
	p.str = amount
	p.Send(packetPlayerStatChange(true, constant.StrID, int32(amount)))
}

// GiveStr to player
func (p *Player) GiveStr(amount int16) {
	p.SetStr(p.str + amount)
}

// SetDex of player
func (p *Player) SetDex(amount int16) {
	p.dex = amount
	p.Send(packetPlayerStatChange(true, constant.DexID, int32(amount)))
}

// GiveDex to player
func (p *Player) GiveDex(amount int16) {
	p.SetDex(p.dex + amount)
}

// SetInt of player
func (p *Player) SetInt(amount int16) {
	p.intt = amount
	p.Send(packetPlayerStatChange(true, constant.IntID, int32(amount)))
}

// GiveInt to player
func (p *Player) GiveInt(amount int16) {
	p.SetInt(p.intt + amount)
}

// SetLuk of player
func (p *Player) SetLuk(amount int16) {
	p.luk = amount
	p.Send(packetPlayerStatChange(true, constant.LukID, int32(amount)))
}

// GiveLuk to player
func (p *Player) GiveLuk(amount int16) {
	p.SetLuk(p.luk + amount)
}

// SetHP of player
func (p *Player) SetHP(amount int16) {
	p.hp = amount
	p.Send(packetPlayerStatChange(true, constant.HpID, int32(amount)))
}

// GiveHP to player
func (p *Player) GiveHP(amount int16) {
	p.SetHP(p.hp + amount)
	if p.hp < 0 {
		p.SetHP(0)
	}
}

// SetMaxHP of player
func (p *Player) SetMaxHP(amount int16) {
	p.maxHP = amount
	p.Send(packetPlayerStatChange(true, constant.MaxHpID, int32(amount)))
}

// SetMP of player
func (p *Player) SetMP(amount int16) {
	p.mp = amount
	p.Send(packetPlayerStatChange(true, constant.MpID, int32(amount)))
}

// GiveMP to player
func (p *Player) GiveMP(amount int16) {
	p.SetMP(p.mp + amount)
	if p.mp < 0 {
		p.SetMP(0)
	}
}

// SetMaxMP of player
func (p *Player) SetMaxMP(amount int16) {
	p.maxMP = amount
	p.Send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

// SetFame of player
func (p *Player) SetFame(amount int16) {

}

// SetGuild of player
func (p *Player) SetGuild(name string, inst instance) {

}

// SetEquipSlotSize of player
func (p *Player) SetEquipSlotSize(size byte) {

}

// SetUseSlotSize of player
func (p *Player) SetUseSlotSize(size byte) {

}

// SetSetUpSlotSize of player
func (p *Player) SetSetUpSlotSize(size byte) {

}

// SetEtcSlotSize of player
func (p *Player) SetEtcSlotSize(size byte) {

}

// SetCashSlotSize of player
func (p *Player) SetCashSlotSize(size byte) {

}

// SetMesos of player
func (p *Player) SetMesos(amount int32) {
	p.mesos = amount
	p.Send(packetPlayerStatChange(false, constant.MesosID, amount))
}

// GiveMesos to player
func (p *Player) GiveMesos(amount int32) {
	p.SetMesos(p.mesos + amount)
}

// SetMiniGameWins of player
func (p *Player) SetMiniGameWins(v int32) {
	p.miniGameWins = v
}

// SetMiniGameLoss of player
func (p *Player) SetMiniGameLoss(v int32) {
	p.miniGameLoss = v
}

// SetMiniGameDraw of player
func (p *Player) SetMiniGameDraw(v int32) {
	p.miniGameDraw = v
}

type movementFrag interface {
	X() int16
	Y() int16
	Foothold() int16
	Stance() byte
}

// UpdateMovement - update player from position data
func (p *Player) UpdateMovement(frag movementFrag) {
	p.pos.SetX(frag.X())
	p.pos.SetY(frag.Y())
	p.foothold = frag.Foothold()
	p.stance = frag.Stance()
}

// SetPos of player
func (p *Player) SetPos(pos pos) {
	p.pos = pos
}

// CheckPos - checks player is within a certain range of a position
func (p Player) CheckPos(pos pos, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = p.pos.X() == pos.X()
	} else {
		xValid = (pos.X()-xRange < p.pos.X() && p.pos.X() < pos.X()+xRange)
	}

	if yRange == 0 {
		xValid = p.pos.Y() == pos.Y()
	} else {
		yValid = (pos.Y()-yRange < p.pos.Y() && p.pos.Y() < pos.Y()+yRange)
	}

	return xValid && yValid
}

// SetFoothold of player
func (p *Player) SetFoothold(fh int16) {
	p.foothold = fh
}

// SetMapID of player
func (p *Player) SetMapID(id int32) {
	p.mapID = id
}

// SetMapPosID of player
func (p *Player) SetMapPosID(pos byte) {
	p.mapPos = pos
}

// GiveItem to player
func (p *Player) GiveItem(newItem item) error {
	findFirstEmptySlot := func(items []item, size byte) (int16, error) {
		slotsUsed := make([]bool, size)

		for _, v := range items {
			if v.SlotID() > 0 {
				slotsUsed[v.SlotID()-1] = true
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

	switch newItem.InvID() {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(p.equip, p.equipSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		newItem.SetAmount(1) // just in case
		p.equip = append(p.equip, newItem)
		p.Send(packetInventoryAddItem(newItem, true))
	case 2: // Use
		var slotID int16
		var index int
		for i, v := range p.use {
			if v.ID() == newItem.ID() && v.Amount() < constant.MaxItemStack {
				slotID = v.SlotID()
				index = i
				break
			}
		}

		if slotID == 0 {
			slotID, err := findFirstEmptySlot(p.use, p.useSlotSize)

			if err != nil {
				return err
			}

			newItem.SetSlotID(slotID)
			p.use = append(p.use, newItem)
			p.Send(packetInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.Amount() - (constant.MaxItemStack - p.use[index].Amount())

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(p.use, p.useSlotSize)

				if err != nil {
					return err
				}

				newItem.SetAmount(remainder)
				newItem.SetSlotID(slotID)
				p.use = append(p.use, newItem)
				p.use[index].SetAmount(constant.MaxItemStack)

				p.Send(packetInventoryAddItems([]item{p.use[index], newItem}, []bool{false, true}))
			} else { // full merge
				p.use[index].SetAmount(p.use[index].Amount() + newItem.Amount())
				p.Send(packetInventoryAddItem(p.use[index], false))
			}
		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(p.setUp, p.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		p.setUp = append(p.setUp, newItem)
		p.Send(packetInventoryAddItem(newItem, true))
	case 4: // Etc
		var slotID int16
		var index int
		for i, v := range p.etc {
			if v.ID() == newItem.ID() && v.Amount() < constant.MaxItemStack {
				slotID = v.SlotID()
				index = i
				break
			}
		}

		if slotID == 0 {
			slotID, err := findFirstEmptySlot(p.etc, p.etcSlotSize)

			if err != nil {
				return err
			}

			newItem.SetSlotID(slotID)
			p.etc = append(p.etc, newItem)
			p.Send(packetInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.Amount() - (constant.MaxItemStack - p.etc[index].Amount())

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(p.etc, p.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.SetAmount(remainder)
				newItem.SetSlotID(slotID)
				p.etc = append(p.etc, newItem)
				p.etc[index].SetAmount(constant.MaxItemStack)

				p.Send(packetInventoryAddItems([]item{p.etc[index], newItem}, []bool{false, true}))
			} else { // full merge
				p.etc[index].SetAmount(p.etc[index].Amount() + newItem.Amount())
				p.Send(packetInventoryAddItem(p.etc[index], false))
			}
		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(p.cash, p.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		p.cash = append(p.cash, newItem)
		p.Send(packetInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.InvID())
	}
	return nil
}

// TakeItem from player
func (p *Player) TakeItem(itemID int32, amount int16) (item, error) {
	return nil, nil
}

// RemoveItem from player
func (p *Player) RemoveItem(remove item) {
	// TODO(Hucaru): change function signature to (id int32, count int16) (invID, slotID, error)

	// findIndex := func(items []item, item item) int {
	// 	for i, v := range items {
	// 		if v.uuid == remove.uuid {
	// 			return i
	// 		}
	// 	}

	// 	return 0
	// }

	// switch remove.invID {
	// case 1:
	// 	if i := findIndex(p.inventory.equip, remove); i != 0 {
	// 		p.inventory.equip[i] = p.inventory.equip[len(p.inventory.equip)-1]
	// 		p.inventory.equip = p.inventory.equip[:len(p.inventory.equip)-1]
	// 	}
	// case 2:
	// 	if i := findIndex(p.inventory.use, remove); i != 0 {
	// 		p.inventory.use[i] = p.inventory.use[len(p.inventory.use)-1]
	// 		p.inventory.use = p.inventory.use[:len(p.inventory.use)-1]
	// 	}
	// case 3:
	// 	if i := findIndex(p.inventory.setUp, remove); i != 0 {
	// 		p.inventory.setUp[i] = p.inventory.setUp[len(p.inventory.setUp)-1]
	// 		p.inventory.setUp = p.inventory.setUp[:len(p.inventory.setUp)-1]
	// 	}
	// case 4:
	// 	if i := findIndex(p.inventory.etc, remove); i != 0 {
	// 		p.inventory.etc[i] = p.inventory.etc[len(p.inventory.etc)-1]
	// 		p.inventory.etc = p.inventory.etc[:len(p.inventory.etc)-1]
	// 	}
	// case 5:
	// 	if i := findIndex(p.inventory.cash, remove); i != 0 {
	// 		p.inventory.cash[i] = p.inventory.cash[len(p.inventory.cash)-1]
	// 		p.inventory.cash = p.inventory.cash[:len(p.inventory.cash)-1]
	// 	}
	// }
}

// GetItem from player
func (p Player) GetItem(invID byte, slotID int16) (item, error) {
	var result item
	var err error

	findItem := func(items []item, slotID int16) (item, error) {
		for _, v := range items {
			if v.SlotID() == slotID {
				return v, nil
			}
		}

		return nil, fmt.Errorf("Unable to get item")
	}

	switch invID {
	case 1:
		result, err = findItem(p.equip, slotID)
	case 2:
		result, err = findItem(p.use, slotID)
	case 3:
		result, err = findItem(p.setUp, slotID)
	case 4:
		result, err = findItem(p.etc, slotID)
	case 5:
		result, err = findItem(p.cash, slotID)
	}

	return result, err
}

// UpdateItem with the same database id
func (p *Player) UpdateItem(orig, new item) {
	var items []item

	switch new.InvID() {
	case 1:
		items = p.equip
	case 2:
		items = p.use
	case 3:
		items = p.setUp
	case 4:
		items = p.etc
	case 5:
		items = p.cash
	}

	for i, v := range items {
		if v.DbID() == new.DbID() {
			items[i] = new
			break
		}
	}
}

// UpdateSkill map entry
func (p *Player) UpdateSkill(updatedSkill Skill) {
	p.skills[updatedSkill.ID] = updatedSkill
	p.Send(packetPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
}

// ID of player
func (p Player) ID() int32 { return p.id }

// AccountID of player
func (p Player) AccountID() int32 { return p.accountID }

// WorldID of player
func (p Player) WorldID() byte { return p.worldID }

// MapID of player
func (p Player) MapID() int32 { return p.mapID }

// MapPos of player
func (p Player) MapPos() byte { return p.mapPos }

// PreviousMap the player was on
func (p Player) PreviousMap() int32 { return p.previousMap }

// PortalCount of player, used in detecting warp hacking
func (p Player) PortalCount() byte { return p.portalCount }

// Job of player
func (p Player) Job() int16 { return p.job }

// Level of player
func (p Player) Level() byte { return p.level }

// Str of player
func (p Player) Str() int16 { return p.str }

//Dex of player
func (p Player) Dex() int16 { return p.dex }

// Int of player
func (p Player) Int() int16 { return p.intt }

// Luk of player
func (p Player) Luk() int16 { return p.luk }

// HP of player
func (p Player) HP() int16 { return p.hp }

// MaxHP of player
func (p Player) MaxHP() int16 { return p.maxHP }

// MP of player
func (p Player) MP() int16 { return p.mp }

// MaxMP of player
func (p Player) MaxMP() int16 { return p.maxMP }

// AP of player
func (p Player) AP() int16 { return p.ap }

// SP of player
func (p Player) SP() int16 { return p.sp }

// Exp of player
func (p Player) Exp() int32 { return p.exp }

// Fame of player
func (p Player) Fame() int16 { return p.fame }

// Name of player
func (p Player) Name() string { return p.name }

// Gender of player
func (p Player) Gender() byte { return p.gender }

// Skin id of player
func (p Player) Skin() byte { return p.skin }

// Face id of player
func (p Player) Face() int32 { return p.face }

// Hair id of player
func (p Player) Hair() int32 { return p.hair }

// ChairID of the chair the player is sitting on
func (p Player) ChairID() int32 { return p.chairID }

// Stance id
func (p Player) Stance() byte { return p.stance }

// Pos of player
func (p Player) Pos() pos { return p.pos }

// Foothold player is currently tied to
func (p Player) Foothold() int16 { return p.foothold }

// Guild name player is currenty part of
func (p Player) Guild() string { return p.guild }

// EquipSlotSize in inventory
func (p Player) EquipSlotSize() byte { return p.equipSlotSize }

// UseSlotSize in inventory
func (p Player) UseSlotSize() byte { return p.useSlotSize }

// SetupSlotSize in inventory
func (p Player) SetupSlotSize() byte { return p.setupSlotSize }

// EtcSlotSize in inventory
func (p Player) EtcSlotSize() byte { return p.etcSlotSize }

//CashSlotSize in inventory
func (p Player) CashSlotSize() byte { return p.cashSlotSize }

// Mesos player currently has
func (p Player) Mesos() int32 { return p.mesos }

// Skills and their levels the player currently has
func (p Player) Skills() map[int32]Skill { return p.skills }

// MiniGameWins between omok and memory
func (p Player) MiniGameWins() int32 { return p.miniGameWins }

// MiniGameDraw betweeen omok and memory
func (p Player) MiniGameDraw() int32 { return p.miniGameDraw }

// MiniGameLoss between omok and memory
func (p Player) MiniGameLoss() int32 { return p.miniGameLoss }

// DisplayBytes used in packets for displaying player in various situations e.g. in field, in mini game room
func (p Player) DisplayBytes() []byte {
	pkt := mpacket.NewPacket()
	pkt.WriteByte(p.gender)
	pkt.WriteByte(p.skin)
	pkt.WriteInt32(p.face)
	pkt.WriteByte(0x00) // ?
	pkt.WriteInt32(p.hair)

	cashWeapon := int32(0)

	for _, b := range p.equip {
		if b.SlotID() < 0 && b.SlotID() > -20 {
			pkt.WriteByte(byte(math.Abs(float64(b.SlotID()))))
			pkt.WriteInt32(b.ID())
		}
	}

	for _, b := range p.equip {
		if b.SlotID() < -100 {
			if b.SlotID() == -111 {
				cashWeapon = b.ID()
			} else {
				pkt.WriteByte(byte(math.Abs(float64(b.SlotID() + 100))))
				pkt.WriteInt32(b.ID())
			}
		}
	}

	pkt.WriteByte(0xFF)
	pkt.WriteByte(0xFF)
	pkt.WriteInt32(cashWeapon)

	return pkt
}

// Save player detail that is not saved via actions
func (p Player) Save(db *sql.DB, inst instance) error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=? WHERE id=?`

	// need to calculate nearest spawn point for mapPos
	portal, err := inst.CalculateNearestSpawnPortal(p.pos)

	if err == nil {
		p.mapPos = portal.ID()
	}

	_, err = db.Exec(query,
		p.skin, p.hair, p.face, p.level, p.job, p.str, p.dex, p.intt, p.luk, p.hp, p.maxHP, p.mp,
		p.maxMP, p.ap, p.sp, p.exp, p.fame, p.mapID, p.mapPos, p.mesos, p.id)

	// TODO: Move these out into relevant item operations, add item, move item etc
	// send sql queries to a dedicated green thread for item updates once a db id is acquired
	for _, v := range p.equip {
		v.Save(db, p.id)
	}

	for _, v := range p.use {
		v.Save(db, p.id)
	}

	for _, v := range p.setUp {
		v.Save(db, p.id)
	}

	for _, v := range p.etc {
		v.Save(db, p.id)
	}

	for _, v := range p.cash {
		v.Save(db, p.id)
	}

	// TODO: Move this into skill book update, this happens 3 times every level (or 15 at a time for min maxers)
	// There has to be a better way of doing this in mysql
	for skillID, skill := range p.skills {
		query = `UPDATE skills SET level=?, cooldown=? WHERE skillID=? AND characterID=?`
		result, err := db.Exec(query, skill.Level, skill.Cooldown, skillID, p.id)

		if rows, _ := result.RowsAffected(); rows < 1 || err != nil {
			query = `INSERT INTO skills (characterID, skillID, level, cooldown) VALUES (?, ?, ?, ?)`
			_, err = db.Exec(query, p.id, skillID, skill.Level, 0)
		}
	}

	return err
}
