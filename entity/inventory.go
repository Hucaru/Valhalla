package entity

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
)

type Inventory struct {
	equip []item
	use   []item
	setUp []item
	etc   []item
	cash  []item
}

// Note: This is a slow way of doing it, save on item acquire, drop, scroll trade etc instead of bulk
func (i Inventory) Save(db *sql.DB, charID int32) {
	save := func(v item) int64 {
		if v.dbID == 0 {
			props := `characterID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,
				str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,
				expireTime,creatorName`

			query := "INSERT into items (" + props + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

			res, err := db.Exec(query,
				charID, v.invID, v.itemID, v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
				v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
				v.expireTime, v.creatorName)

			if err != nil {
				log.Println(err)
				return 0
			}

			v.dbID, err = res.LastInsertId()

			if err != nil {
				log.Println(err)
				return 0
			}
		} else {
			props := `slotNumber=?,amount=?,flag=?,upgradeSlots=?,level=?,
			str=?,dex=?,intt=?,luk=?,hp=?,mp=?,watk=?,matk=?,wdef=?,mdef=?,accuracy=?,avoid=?,hands=?,speed=?,jump=?,
			expireTime=?`

			query := "UPDATE items SET " + props + " WHERE id=?"

			_, err := db.Exec(query,
				v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
				v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
				v.expireTime, v.dbID)

			if err != nil {
				log.Println(err)
				return 0
			}
		}

		return v.dbID
	}

	for j, v := range i.equip {
		i.equip[j].dbID = save(v)
	}

	for j, v := range i.use {
		i.use[j].dbID = save(v)
	}

	for j, v := range i.setUp {
		i.setUp[j].dbID = save(v)
	}

	for j, v := range i.etc {
		i.etc[j].dbID = save(v)
	}

	for j, v := range i.cash {
		i.cash[j].dbID = save(v)
	}
}

func (i *Inventory) addEquip(newItem item) {
	i.equip = append(i.equip, newItem)
}

func getInventoryFromCharID(db *sql.DB, id int32) Inventory {
	filter := "id,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := db.Query("SELECT "+filter+" FROM items WHERE characterID=?", id)

	if err != nil {
		panic(err)
	}

	inventory := Inventory{}

	defer row.Close()

	for row.Next() {

		item := item{uuid: uuid.New()}

		row.Scan(&item.dbID,
			&item.invID,
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

		item.calculateWeaponType()

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
