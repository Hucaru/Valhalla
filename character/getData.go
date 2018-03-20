package character

import (
	"sync"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/inventory"
)

func GetCharacterSkills(charID uint32) map[uint32]uint32 {
	filter := "skillID,level"
	row, err := connection.Db.Query("SELECT "+filter+" FROM skills WHERE characterID=?", charID)

	if err != nil {
		panic(err.Error())
	}

	defer row.Close()

	skills := make(map[uint32]uint32)

	for row.Next() {
		var ID uint32
		var level uint32

		row.Scan(&ID, &level)

		skills[ID] = level
	}

	return skills
}

func GetCharacterItems(charID uint32) []inventory.Item {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := connection.Db.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []inventory.Item

	defer row.Close()

	for row.Next() {

		var item inventory.Item
		var invID byte
		var itemID uint32
		var slotID int16
		var amount uint16
		var flag uint16
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

		newChar.mutex = &sync.RWMutex{}

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
