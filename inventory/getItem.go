package inventory

import "github.com/Hucaru/Valhalla/connection"

func GetCharacterInventory(charID int32) []Item {
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
		item.SetSlotID(slotID)
		item.SetSlotID(slotID)
		item.SetAmount(amount)
		item.SetFlag(flag)
		item.SetUpgradeSlots(upgradeSlots)
		item.SetReqLevel(level)
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

func GetCharacterStorage(charID int32) {

}
