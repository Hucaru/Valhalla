package character

import (
	"github.com/Hucaru/Valhalla/common/connection"
)

// Character struct
type Character struct {
	CharID          uint32
	UserID          uint32
	WorldID         uint32
	Name            string
	Gender          byte
	Skin            byte
	Face            uint32
	Hair            uint32
	Level           byte
	Job             uint16
	Str             uint16
	Dex             uint16
	Intt            uint16
	Luk             uint16
	HP              uint16
	MaxHP           uint16
	MP              uint16
	MaxMP           uint16
	AP              uint16
	SP              uint16
	EXP             uint32
	Fame            uint16
	CurrentMap      uint32
	CurrentMapPos   byte
	PreviousMap     uint32
	FeeMarketReturn uint32
	Mesos           uint32
	EquipSlotSize   byte
	UsetSlotSize    byte
	SetupSlotSize   byte
	EtcSlotSize     byte
	CashSlotSize    byte

	Equips []Equip
}

// Items struct
type Equip struct {
	ItemID       uint32
	SlotID       int32
	UpgradeSlots byte
	Level        byte
	Str          uint16
	Dex          uint16
	Intt         uint16
	Luk          uint16
	HP           uint16
	MP           uint16
	Watk         uint16
	Matk         uint16
	Wdef         uint16
	Mdef         uint16
	Accuracy     uint16
	Avoid        uint16
	Hands        uint16
	Speed        uint16
	Jump         uint16
	ExpireTime   uint64
	OwnerName    string
}

func GetCharacter(charID uint32) Character {
	var newChar Character
	filter := "id,userID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	err := connection.Db.QueryRow("SELECT "+filter+" FROM characters where id=?", charID).Scan(&newChar.CharID,
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
		&newChar.EquipSlotSize,
		&newChar.UsetSlotSize,
		&newChar.SetupSlotSize,
		&newChar.EtcSlotSize,
		&newChar.CashSlotSize)

	if err != nil {
		panic(err)
	}

	return newChar
}

func GetCharacterItems(charID uint32) []Equip {
	filter := "itemID,slotNumber,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,ownerName"
	row, err := connection.Db.Query("SELECT "+filter+" FROM equips WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	var items []Equip

	defer row.Close()

	for row.Next() {
		var item Equip

		row.Scan(&item.ItemID,
			&item.SlotID,
			&item.UpgradeSlots,
			&item.Level,
			&item.Str,
			&item.Dex,
			&item.Intt,
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
			&item.OwnerName)

		items = append(items, item)
	}

	return items
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
			&newChar.EquipSlotSize,
			&newChar.UsetSlotSize,
			&newChar.SetupSlotSize,
			&newChar.EtcSlotSize,
			&newChar.CashSlotSize)

		if err != nil {
			panic(err)
		}

		equips, err := connection.Db.Query("SELECT itemID, slotNumber FROM equips WHERE characterID=?", newChar.CharID)

		if err != nil {
			panic(err)
		}

		defer equips.Close()

		for equips.Next() {
			var equip Equip

			equips.Scan(&equip.ItemID, &equip.SlotID)
			newChar.Equips = append(newChar.Equips, equip)
		}

		characters = append(characters, newChar)
	}

	return characters
}
