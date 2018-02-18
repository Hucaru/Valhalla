package character

import (
	"sync"

	"github.com/Hucaru/Valhalla/common/connection"
)

func GetCharacterSkills(charID uint32) []Skill {
	filter := "skillID,level"
	row, err := connection.Db.Query("SELECT "+filter+" FROM skills WHERE characterID=?", charID)

	if err != nil {
		panic(err.Error())
	}

	defer row.Close()

	var skills []Skill

	for row.Next() {
		var newSkill Skill

		var ID uint32
		var level byte

		row.Scan(&ID, &level)

		newSkill.SetID(ID)
		newSkill.SetLevel(level)

		skills = append(skills, newSkill)
	}

	return skills
}

func GetCharacterItems(charID uint32) []Item {
	filter := "inventoryID,itemID,slotNumber,amount,flag,creatorName,expiration"
	row, err := connection.Db.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []Item

	defer row.Close()

	for row.Next() {
		var item Item

		var invID byte
		var slotNumber byte
		var itemID uint32
		var expiration uint64
		var amount uint16
		var creatorName string
		var flag uint16

		row.Scan(&invID, &itemID, &slotNumber, &amount, &flag, &creatorName, &expiration)

		item.SetInvID(invID)
		item.SetItemID(itemID)
		item.SetSlotNumber(slotNumber)
		item.SetExpiration(expiration)
		item.SetAmount(amount)
		item.SetCreatorName(creatorName)
		item.SetFlag(flag)

		items = append(items, item)
	}

	return items
}

func GetCharacterEquips(charID uint32) []Equip {
	filter := "itemID,slotNumber,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := connection.Db.Query("SELECT "+filter+" FROM equips WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []Equip

	defer row.Close()

	for row.Next() {

		var item Equip

		var itemID uint32
		var slotID int32
		var upgradeSlots byte
		var level byte
		var str uint16
		var dex uint16
		var intt uint16
		var luk uint16
		var hp uint16
		var mp uint16
		var watk uint16
		var matk uint16
		var wdef uint16
		var mdef uint16
		var accuracy uint16
		var avoid uint16
		var hands uint16
		var speed uint16
		var jump uint16
		var expireTime uint64
		var creatorName string

		row.Scan(&itemID,
			&slotID,
			&upgradeSlots,
			&level,
			&str,
			&dex,
			&intt,
			&luk,
			&hp,
			&mp,
			&watk,
			&matk,
			&wdef,
			&mdef,
			&accuracy,
			&avoid,
			&hands,
			&speed,
			&jump,
			&expireTime,
			&creatorName)

		item.SetItemID(itemID)
		item.SetSlotID(slotID)
		item.SetUpgradeSlots(upgradeSlots)
		item.SetLevel(level)
		item.SetStr(str)
		item.SetDex(dex)
		item.SetInt(intt)
		item.SetLuk(luk)
		item.SetHP(hp)
		item.SetMP(mp)
		item.SetWatk(watk)
		item.SetMatk(matk)
		item.SetWdef(wdef)
		item.SetMdef(mdef)
		item.SetAccuracy(accuracy)
		item.SetAvoid(avoid)
		item.SetHands(hands)
		item.SetSpeed(speed)
		item.SetJump(jump)
		item.SetExpireTime(expireTime)
		item.SetCreatorName(creatorName)

		items = append(items, item)
	}

	return items
}

func GetCharacter(charID uint32) Character {
	var newChar Character
	filter := "id,userID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	var userID uint32
	var worldID uint32
	var name string
	var gender byte
	var skin byte
	var face uint32
	var hair uint32
	var level byte
	var job uint16
	var str uint16
	var dex uint16
	var intt uint16
	var luk uint16
	var hp uint16
	var maxHP uint16
	var mp uint16
	var maxMP uint16
	var ap uint16
	var sp uint16
	var exp uint32
	var fame uint16
	var currentMap uint32
	var currentMapPos byte
	var previousMap uint32
	var feeMarketReturn uint32
	var mesos uint32
	var equipSlotSize byte
	var useSlotSize byte
	var setupSlotSize byte
	var etcSlotSize byte
	var cashSlotSize byte

	err := connection.Db.QueryRow("SELECT "+filter+" FROM characters where id=?", charID).Scan(&charID,
		&userID,
		&worldID,
		&name,
		&gender,
		&skin,
		&hair,
		&face,
		&level,
		&job,
		&str,
		&dex,
		&intt,
		&luk,
		&hp,
		&maxHP,
		&mp,
		&maxMP,
		&ap,
		&sp,
		&exp,
		&fame,
		&currentMap,
		&currentMapPos,
		&previousMap,
		&mesos,
		&equipSlotSize,
		&useSlotSize,
		&setupSlotSize,
		&etcSlotSize,
		&cashSlotSize)

	if err != nil {
		panic(err)
	}

	newChar.mutex = &sync.Mutex{}

	newChar.SetCharID(charID)
	newChar.SetUserID(userID)
	newChar.SetWorldID(worldID)
	newChar.SetName(name)
	newChar.SetGender(gender)
	newChar.SetSkin(skin)
	newChar.SetFace(face)
	newChar.SetHair(hair)
	newChar.SetLevel(level)
	newChar.SetJob(job)
	newChar.SetStr(str)
	newChar.SetDex(dex)
	newChar.SetInt(intt)
	newChar.SetLuk(luk)
	newChar.SetHP(hp)
	newChar.SetMaxHP(maxHP)
	newChar.SetMP(mp)
	newChar.SetMaxMp(maxMP)
	newChar.SetAP(ap)
	newChar.SetSP(sp)
	newChar.SetEXP(exp)
	newChar.SetFame(fame)
	newChar.SetCurrentMap(currentMap)
	newChar.SetCurrentMapPos(currentMapPos)
	newChar.SetPreviousMap(previousMap)
	newChar.SetFreeMarketReturn(feeMarketReturn)
	newChar.SetMesos(mesos)
	newChar.SetEquipSlotSize(equipSlotSize)
	newChar.SetUseSlotSize(useSlotSize)
	newChar.SetSetupSlotSize(setupSlotSize)
	newChar.SetEtcSlotSize(etcSlotSize)
	newChar.SetCashSlotSize(cashSlotSize)

	return newChar
}

func GetCharacters(userID uint32, worldID uint32) []Character {
	filter := "id,userID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	chars, err := connection.Db.Query("SELECT "+filter+" FROM characters WHERE userID=? AND worldID=?", userID, worldID)

	if err != nil {
		panic(err)
	}

	defer chars.Close()

	var characters []Character

	for chars.Next() {
		var newChar Character

		newChar.mutex = &sync.Mutex{}

		var charID uint32
		var userID uint32
		var worldID uint32
		var name string
		var gender byte
		var skin byte
		var face uint32
		var hair uint32
		var level byte
		var job uint16
		var str uint16
		var dex uint16
		var intt uint16
		var luk uint16
		var hp uint16
		var maxHP uint16
		var mp uint16
		var maxMP uint16
		var ap uint16
		var sp uint16
		var exp uint32
		var fame uint16
		var currentMap uint32
		var currentMapPos byte
		var previousMap uint32
		var feeMarketReturn uint32
		var mesos uint32
		var equipSlotSize byte
		var useSlotSize byte
		var setupSlotSize byte
		var etcSlotSize byte
		var cashSlotSize byte

		err = chars.Scan(&charID,
			&userID,
			&worldID,
			&name,
			&gender,
			&skin,
			&hair,
			&face,
			&level,
			&job,
			&str,
			&dex,
			&intt,
			&luk,
			&hp,
			&maxHP,
			&mp,
			&maxMP,
			&ap,
			&sp,
			&exp,
			&fame,
			&currentMap,
			&currentMapPos,
			&previousMap,
			&mesos,
			&equipSlotSize,
			&useSlotSize,
			&setupSlotSize,
			&etcSlotSize,
			&cashSlotSize)

		if err != nil {
			panic(err)
		}

		newChar.SetCharID(charID)
		newChar.SetUserID(userID)
		newChar.SetWorldID(worldID)
		newChar.SetName(name)
		newChar.SetGender(gender)
		newChar.SetSkin(skin)
		newChar.SetFace(face)
		newChar.SetHair(hair)
		newChar.SetLevel(level)
		newChar.SetJob(job)
		newChar.SetStr(str)
		newChar.SetDex(dex)
		newChar.SetInt(intt)
		newChar.SetLuk(luk)
		newChar.SetHP(hp)
		newChar.SetMaxHP(maxHP)
		newChar.SetMP(mp)
		newChar.SetMaxMp(maxMP)
		newChar.SetAP(ap)
		newChar.SetSP(sp)
		newChar.SetEXP(exp)
		newChar.SetFame(fame)
		newChar.SetCurrentMap(currentMap)
		newChar.SetCurrentMapPos(currentMapPos)
		newChar.SetPreviousMap(previousMap)
		newChar.SetFreeMarketReturn(feeMarketReturn)
		newChar.SetMesos(mesos)
		newChar.SetEquipSlotSize(equipSlotSize)
		newChar.SetUseSlotSize(useSlotSize)
		newChar.SetSetupSlotSize(setupSlotSize)
		newChar.SetEtcSlotSize(etcSlotSize)
		newChar.SetCashSlotSize(cashSlotSize)

		equips, err := connection.Db.Query("SELECT itemID, slotNumber FROM equips WHERE characterID=?", newChar.GetCharID())

		if err != nil {
			panic(err)
		}

		defer equips.Close()

		var equipment []Equip

		for equips.Next() {
			var equip Equip

			var itemID uint32
			var slotID int32

			equips.Scan(&itemID, &slotID)

			equip.SetItemID(itemID)
			equip.SetSlotID(slotID)

			equipment = append(equipment, equip)
		}

		newChar.SetEquips(equipment)

		characters = append(characters, newChar)
	}

	return characters
}
