package character

import (
	"sync"

	"github.com/Hucaru/Valhalla/connection"
)

func GetCharacterSkills(charID int32) map[int32]int32 {
	filter := "skillID,level"
	row, err := connection.Db.Query("SELECT "+filter+" FROM skills WHERE characterID=?", charID)

	if err != nil {
		panic(err.Error())
	}

	defer row.Close()

	skills := make(map[int32]int32)

	for row.Next() {
		var ID int32
		var level int32

		row.Scan(&ID, &level)

		skills[ID] = level
	}

	return skills
}

func GetCharacterItems(charID int32) []Item {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := connection.Db.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []Item

	defer row.Close()

	for row.Next() {

		var item Item
		var invID byte
		var itemID int32
		var slotID int16
		var amount int16
		var flag int16
		var upgradeSlots byte
		var level byte
		var str int16
		var dex int16
		var intt int16
		var luk int16
		var hp int16
		var mp int16
		var watk int16
		var matk int16
		var wdef int16
		var mdef int16
		var accuracy int16
		var avoid int16
		var hands int16
		var speed int16
		var jump int16
		var expireTime uint64
		var creatorName string

		row.Scan(&invID,
			&itemID,
			&slotID,
			&amount,
			&flag,
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

		item.SetInvID(invID)
		item.SetItemID(itemID)
		item.SetSlotNumber(slotID)
		item.SetAmount(amount)
		item.SetFlag(flag)
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
		item.SetExpirationTime(expireTime)
		item.SetCreatorName(creatorName)

		items = append(items, item)
	}

	return items
}

func GetCharacter(charID int32) Character {
	var newChar Character
	filter := "id,userID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	var userID int32
	var worldID int32
	var name string
	var gender byte
	var skin byte
	var face int32
	var hair int32
	var level byte
	var job int16
	var str int16
	var dex int16
	var intt int16
	var luk int16
	var hp int16
	var maxHP int16
	var mp int16
	var maxMP int16
	var ap int16
	var sp int16
	var exp int32
	var fame int16
	var currentMap int32
	var currentMapPos byte
	var previousMap int32
	var feeMarketReturn int32
	var mesos int32
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

	newChar.mutex = &sync.RWMutex{}

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
	newChar.SetMaxMP(maxMP)
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

	newChar.SetChairID(0)

	return newChar
}

func GetCharacters(userID int32, worldID int32) []Character {
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

		newChar.mutex = &sync.RWMutex{}

		var charID int32
		var userID int32
		var worldID int32
		var name string
		var gender byte
		var skin byte
		var face int32
		var hair int32
		var level byte
		var job int16
		var str int16
		var dex int16
		var intt int16
		var luk int16
		var hp int16
		var maxHP int16
		var mp int16
		var maxMP int16
		var ap int16
		var sp int16
		var exp int32
		var fame int16
		var currentMap int32
		var currentMapPos byte
		var previousMap int32
		var feeMarketReturn int32
		var mesos int32
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
		newChar.SetMaxMP(maxMP)
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

		newChar.SetItems(GetCharacterItems(charID))

		characters = append(characters, newChar)
	}

	return characters
}
