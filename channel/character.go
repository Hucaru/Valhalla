package channel

import (
	"log"
	"math/rand"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

type MapleCharacter struct {
	character.Character
	conn interop.ClientConn // Might be worth compositing this in?
}

func (c *MapleCharacter) SendPacket(p maplepacket.Packet) {
	if len(p) > 0 {
		c.conn.Write(p)
	}
}

func (c *MapleCharacter) GetConn() interop.ClientConn {
	return c.conn
}

func (c *MapleCharacter) IsAdmin() bool {
	return c.conn.IsAdmin()
}

func (c *MapleCharacter) SetHP(hp uint16) {
	c.Character.SetHP(c.GetHP() + hp)

	if c.GetHP() > c.GetMaxHP() {
		c.Character.SetHP(c.GetMaxHP())
	}

	c.conn.Write(packets.PlayerStatChange(true, constants.HP_ID, uint32(c.GetHP())))
}

func (c *MapleCharacter) SetMP(mp uint16) {
	c.Character.SetMP(c.Character.GetMP() + mp)

	if c.GetMP() > c.GetMaxMP() {
		c.Character.SetMP(c.GetMaxMP())
	}

	c.conn.Write(packets.PlayerStatChange(true, constants.MP_ID, uint32(c.GetMP())))
}

func (c *MapleCharacter) SetAP(ap uint16) {
	c.Character.SetAP(ap)
	c.conn.Write(packets.PlayerStatChange(true, constants.AP_ID, uint32(ap)))
}

func (c *MapleCharacter) SetStr(str uint16) {
	var maxValue uint16 = 2000

	if c.GetStr() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetStr(str)

	c.conn.Write(packets.PlayerStatChange(true, constants.STR_ID, uint32(str)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetDex(dex uint16) {
	var maxValue uint16 = 2000

	if c.GetDex() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetDex(dex)

	c.conn.Write(packets.PlayerStatChange(true, constants.DEX_ID, uint32(dex)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetInt(intt uint16) {
	var maxValue uint16 = 2000

	if c.GetInt() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetInt(intt)

	c.conn.Write(packets.PlayerStatChange(true, constants.INT_ID, uint32(intt)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetLuk(luk uint16) {
	var maxValue uint16 = 2000

	if c.GetLuk() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetLuk(luk)

	c.conn.Write(packets.PlayerStatChange(true, constants.LUK_ID, uint32(luk)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxHP(mp uint16) {
	var maxValue uint16 = 30000

	if c.GetMaxHP() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxHP(mp)

	c.conn.Write(packets.PlayerStatChange(true, constants.MAX_HP_ID, uint32(mp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxMP(hp uint16) {
	var maxValue uint16 = 30000

	if c.GetMaxMP() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxMP(hp)

	c.conn.Write(packets.PlayerStatChange(true, constants.MAX_MP_ID, uint32(hp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetSP(sp uint16) {
	c.Character.SetSP(sp)
	c.conn.Write(packets.PlayerStatChange(true, constants.SP_ID, uint32(sp)))
}

func (c *MapleCharacter) UpdateSkill(id, level uint32) {
	c.Character.UpdateSkill(id, level)
	c.SetSP(c.GetSP() - 1)
	c.conn.Write(packets.PlayerSkillBookUpdate(id, level))
}

func (c *MapleCharacter) ChangeMap(mapID uint32, portal maplePortal, pID byte) {
	Maps.GetMap(c.GetCurrentMap()).RemovePlayer(c.conn)

	c.SetX(portal.GetX())
	c.SetY(portal.GetY())

	c.conn.Write(packets.MapChange(mapID, 1, pID, c.GetHP())) // replace 1 with channel id
	c.SetCurrentMap(mapID)
	Maps.GetMap(mapID).AddPlayer(c.conn)
}

func (c *MapleCharacter) LevelUP() {
	newAP := c.Character.GetAP() + 5
	c.Character.SetAP(newAP)

	newSP := c.Character.GetSP() + 3
	c.Character.SetSP(newSP)

	var hpToAdd uint16
	var mpToAdd uint16

	levelUpHp := func(classIncrease uint16, bonus uint16) uint16 {
		return uint16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease uint16, bonus uint16) uint16 {
		return uint16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch int(c.Character.GetJob() / 100) {
	case 0:
		hpToAdd = levelUpHp(constants.BEGGINNER_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.BEGGINNER_MP_ADD, c.Character.GetInt())
	case 1:
		hpToAdd = levelUpHp(constants.WARRIOR_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.WARRIOR_MP_ADD, c.Character.GetInt())
	case 2:
		hpToAdd = levelUpHp(constants.MAGICIAN_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.MAGICIAN_MP_ADD, 2*c.Character.GetInt())
	case 3:
		hpToAdd = levelUpHp(constants.BOWMAN_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.BOWMAN_MP_ADD, c.Character.GetInt())
	case 4:
		hpToAdd = levelUpHp(constants.THIEF_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.THIEF_MP_ADD, c.Character.GetInt())
	case 5:
		hpToAdd = constants.ADMIN_HP_ADD
		mpToAdd = constants.ADMIN_MP_ADD
	default:
		log.Println("Unknown Job ID:", c.Character.GetJob())
	}

	newHp := c.Character.GetMaxHP() + hpToAdd
	c.Character.SetMaxHP(newHp)
	c.Character.SetHP(newHp)

	newMp := c.Character.GetMaxMP() + mpToAdd
	c.Character.SetMaxMP(newMp)
	c.Character.SetMP(newMp)

	c.conn.Write(packets.PlayerStatChange(false, constants.HP_ID, uint32(newHp)))
	c.conn.Write(packets.PlayerStatChange(false, constants.MAX_HP_ID, uint32(newHp)))

	c.conn.Write(packets.PlayerStatChange(false, constants.MP_ID, uint32(newHp)))
	c.conn.Write(packets.PlayerStatChange(false, constants.MAX_MP_ID, uint32(newHp)))

	c.conn.Write(packets.PlayerStatChange(false, constants.AP_ID, uint32(newAP)))
	c.conn.Write(packets.PlayerStatChange(false, constants.SP_ID, uint32(newSP)))
}

func (c *MapleCharacter) SetLevel(level byte) {
	Maps.GetMap(c.GetCurrentMap()).SendPacket(packets.PlayerLevelUpAnimation(c.GetCharID()))
	delta := level - c.Character.GetLevel()

	if delta > 0 {
		for i := byte(0); i < delta; i++ {
			c.LevelUP()
		}
	}

	c.Character.SetLevel(level)
	c.conn.Write(packets.PlayerStatChange(true, constants.LEVEL_ID, uint32(level)))

}

func (c *MapleCharacter) SetJob(jobID uint16) {
	c.Character.SetJob(jobID)
	c.conn.Write(packets.PlayerStatChange(true, constants.JOB_ID, uint32(jobID)))
}

func (c *MapleCharacter) SetMesos(val uint32) {
	c.Character.SetMesos(val)
	c.conn.Write(packets.PlayerStatChange(true, constants.MESOS_ID, val))
}

func (c *MapleCharacter) GiveMesos(val uint32) {
	c.SetMesos(c.GetMesos() + val)
}

func (c *MapleCharacter) TakeMesos(val uint32) {
	c.SetMesos(c.GetMesos() - val)
}

func (c *MapleCharacter) GiveEXP(val uint32, whiteText, appearInChat bool) {
	var giveEXP func(val uint32)

	giveEXP = func(val uint32) {
		if c.GetLevel() > 199 {
			c.SetEXP(0)
			c.conn.Write(packets.PlayerStatChange(true, constants.EXP_ID, 0))
		} else if c.GetEXP()+val >= ExpTable[c.GetLevel()-1] {
			leftOver := c.GetEXP() + val - ExpTable[c.GetLevel()-1]
			c.SetLevel(c.GetLevel() + 1)
			giveEXP(leftOver) // Recursive call to allow multiple level ups
		} else {
			c.SetEXP(c.GetEXP() + val)
			c.conn.Write(packets.PlayerStatChange(true, constants.EXP_ID, c.GetEXP()))
		}
	}

	giveEXP(val)

	c.conn.Write(packets.MessageExpGained(whiteText, appearInChat, val))
}

func (c *MapleCharacter) TakeEXP(val uint32) {
	if c.GetEXP() < val {
		c.SetEXP(0)
	} else {
		c.SetEXP(c.GetEXP() - val)
	}
}

func (c *MapleCharacter) GiveItem(item character.Item) {
	log.Println("Implement Give item:", item)
}

func (c *MapleCharacter) TakeItem(slotID int16, itemID uint32, ammount uint16) {
	log.Println("Implement take item:", slotID, itemID, ammount)
}

func (c *MapleCharacter) TakeDamage(ammount uint32) {
	delta := int32(c.Character.GetHP()) - int32(ammount)

	var newHp uint16

	if delta < 1 {
		newHp = 0
	} else {
		newHp = uint16(delta)
	}

	c.Character.SetHP(newHp)
	c.conn.Write(packets.PlayerStatChange(false, constants.HP_ID, uint32(newHp)))
}
