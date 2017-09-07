package character

import (
	"github.com/Hucaru/Valhalla/common/connection"
)

// Character struct
type Character struct {
	CharID          int32
	UserID          int32
	WorldID         int32
	Name            string
	Gender          byte
	Skin            byte
	Face            int32
	Hair            int32
	Level           byte
	Job             int16
	Str             int16
	Dex             int16
	Intt            int16
	Luk             int16
	HP              int16
	MaxHP           int16
	MP              int16
	MaxMP           int16
	AP              int16
	SP              int16
	EXP             int32
	Fame            int16
	CurrentMap      int32
	CurrentMapPos   byte
	PreviousMap     int32
	FeeMarketReturn int32
	Mesos           int32

	Items []Items
}

// Items struct
type Items struct {
	ItemID int32
	SlotID int32
}

func GetCharacters(userID uint32, worldID uint32) []Character {
	chars, err := connection.Db.Query("SELECT * FROM characters WHERE userID=? AND worldID=?", userID, worldID)

	if err != nil {
		panic(err)
	}

	defer chars.Close()

	var characters []Character

	for chars.Next() {
		var newChar Character

		err = chars.Scan(&newChar.CharID,
			&newChar.UserID,
			&newChar.WorldID,
			&newChar.Name,
			&newChar.Gender,
			&newChar.Skin,
			&newChar.Hair,
			&newChar.Face,
			&newChar.Level,
			&newChar.Job,
			&newChar.Str,
			&newChar.Dex,
			&newChar.Intt,
			&newChar.Luk,
			&newChar.HP,
			&newChar.MaxHP,
			&newChar.MP,
			&newChar.MaxMP,
			&newChar.AP,
			&newChar.SP,
			&newChar.EXP,
			&newChar.Fame,
			&newChar.CurrentMap,
			&newChar.CurrentMapPos,
			&newChar.PreviousMap,
			&newChar.Mesos,
		)

		if err != nil {
			panic(err)
		}

		items, err := connection.Db.Query("SELECT itemID, slotNumber FROM items WHERE characterID=?", newChar.CharID)

		if err != nil {
			panic(err)
		}

		defer items.Close()

		for items.Next() {
			var item Items

			items.Scan(&item.ItemID, &item.SlotID)
			newChar.Items = append(newChar.Items, item)
		}

		characters = append(characters, newChar)
	}

	return characters
}
