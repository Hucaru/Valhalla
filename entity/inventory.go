package entity

import (
	"database/sql"
)

type Inventory struct {
	equip []item
	use   []item
	setUp []item
	etc   []item
	cash  []item
}

func (i Inventory) save(id int32) {
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

func getInventoryFromCharID(db *sql.DB, id int32) Inventory {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := db.Query("SELECT "+filter+" FROM items WHERE characterID=?", id)

	if err != nil {
		panic(err)
	}

	inventory := Inventory{}

	defer row.Close()

	for row.Next() {

		item := item{}

		row.Scan(&item.invID,
			&item.itemID,
			&item.slotID,
			&item.amount,
			&item.flag,
			&item.upgradeSlots,
			&item.scrollLevel,
			&item.str,
			&item.dex,
			&item.intt,
			&item.luk,
			&item.hp,
			&item.mp,
			&item.watk,
			&item.matk,
			&item.wdef,
			&item.mdef,
			&item.accuracy,
			&item.avoid,
			&item.hands,
			&item.speed,
			&item.jump,
			&item.expireTime,
			&item.creatorName)

		switch item.invID {
		case 1:
			inventory.equip = append(inventory.equip, item)
		case 2:
			inventory.use = append(inventory.use, item)
		case 3:
			inventory.setUp = append(inventory.setUp, item)
		case 4:
			inventory.etc = append(inventory.etc, item)
		case 5:
			inventory.cash = append(inventory.cash, item)
		default:
		}

	}

	return inventory
}
