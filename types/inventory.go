package types

import (
	"github.com/Hucaru/Valhalla/database"
	"github.com/google/uuid"
)

type Inventory struct {
	Equip []Item
	Use   []Item
	SetUp []Item
	Etc   []Item
	Cash  []Item

	Mesos int32
}

func (i Inventory) Save(id int32) {
	// delete all items in inventory

	// save all items in inventory
	// for _, v := range i.Equip {

	// }

	// for _, v := range i.Use {

	// }

	// for _, v := range i.SetUp {

	// }

	// for _, v := range i.Etc {

	// }

	// for _, v := range i.Cash {

	// }
}

func GetInventoryFromCharID(id int32) Inventory {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := database.Handle.Query("SELECT "+filter+" FROM items WHERE characterID=?", id)

	if err != nil {
		panic(err)
	}

	inventory := Inventory{}

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

		switch item.InvID {
		case 1:
			inventory.Equip = append(inventory.Equip, item)
		case 2:
			inventory.Use = append(inventory.Use, item)
		case 3:
			inventory.SetUp = append(inventory.SetUp, item)
		case 4:
			inventory.Etc = append(inventory.Etc, item)
		case 5:
			inventory.Cash = append(inventory.Cash, item)
		default:
		}

	}

	return inventory
}
