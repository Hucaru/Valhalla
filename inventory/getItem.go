package inventory

import (
	"github.com/Hucaru/Valhalla/connection"
	"github.com/google/uuid"
)

func GetCharacterInventory(charID int32) []Item {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := connection.Db.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []Item

	defer row.Close()

	for row.Next() {

		item := Item{UUID: uuid.Must(uuid.NewRandom())}

		row.Scan(&item.InvID,
			&item.ItemID,
			&item.SlotID,
			&item.Amount,
			&item.Flag,
			&item.UpgradeSlots,
			&item.ScrollLevel,
			&item.Str,
			&item.Dex,
			&item.Int,
			&item.Luk,
			&item.HP,
			&item.MP,
			&item.Watk,
			&item.Matk,
			&item.Wdef,
			&item.Mdef,
			&item.Accuracy,
			&item.Avoid,
			&item.Hands,
			&item.Speed,
			&item.Jump,
			&item.ExpireTime,
			&item.CreatorName)

		items = append(items, item)
	}

	return items
}

func GetCharacterStorage(charID int32) {

}
