package types

import (
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/nx"
)

type Character struct {
	ID        int32
	AccountID int32
	WorldID   byte

	CurrentMap    int32
	CurrentMapPos byte
	PreviousMap   int32

	Job int16

	Level byte
	Str   int16
	Dex   int16
	Int   int16
	Luk   int16
	HP    int16
	MaxHP int16
	MP    int16
	MaxMP int16
	AP    int16
	SP    int16
	EXP   int32
	Fame  int16

	Avatar

	EquipSlotSize byte
	UseSlotSize   byte
	SetupSlotSize byte
	EtcSlotSize   byte
	CashSlotSize  byte

	Inventory

	Skills map[int32]int32

	MiniGameWins, MiniGameTies, MiniGameLosses int32
}

func (c Character) Save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mesos=? WHERE id=?`

	// need to calculate nearest spawn point for mapPos

	records, err := database.Handle.Query(query,
		c.Skin, c.Hair, c.Face, c.Level, c.Job, c.Str, c.Dex, c.Int, c.Luk, c.HP, c.MaxHP, c.MP,
		c.MaxMP, c.AP, c.SP, c.EXP, c.Fame, c.CurrentMap, c.Mesos, c.ID)

	defer records.Close()

	c.Inventory.Save(c.ID)

	return err
}

func GetCharactersFromAccountWorldID(accountID int32, worldID byte) []Character {
	c := []Character{}

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	chars, err := database.Handle.Query("SELECT "+filter+" FROM characters WHERE accountID=? AND worldID=?", accountID, worldID)

	if err != nil {
		panic(err)
	}

	defer chars.Close()

	for chars.Next() {
		var char Character

		err = chars.Scan(&char.ID, &char.AccountID, &char.WorldID, &char.Name, &char.Gender, &char.Skin, &char.Hair,
			&char.Face, &char.Level, &char.Job, &char.Str, &char.Dex, &char.Int, &char.Luk, &char.HP, &char.MaxHP,
			&char.MP, &char.MaxMP, &char.AP, &char.SP, &char.EXP, &char.Fame, &char.CurrentMap, &char.CurrentMapPos,
			&char.PreviousMap, &char.Mesos, &char.EquipSlotSize, &char.UseSlotSize, &char.SetupSlotSize,
			&char.EtcSlotSize, &char.CashSlotSize)

		if err != nil {
			panic(err)
		}

		char.Inventory = GetInventoryFromCharID(char.ID)

		c = append(c, char)
	}

	return c
}

func GetCharacterFromID(id int32) Character {
	var char Character

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	err := database.Handle.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&char.ID,
		&char.AccountID, &char.WorldID, &char.Name, &char.Gender, &char.Skin, &char.Hair, &char.Face,
		&char.Level, &char.Job, &char.Str, &char.Dex, &char.Int, &char.Luk, &char.HP, &char.MaxHP, &char.MP,
		&char.MaxMP, &char.AP, &char.SP, &char.EXP, &char.Fame, &char.CurrentMap, &char.CurrentMapPos,
		&char.PreviousMap, &char.Mesos, &char.EquipSlotSize, &char.UseSlotSize, &char.SetupSlotSize,
		&char.EtcSlotSize, &char.CashSlotSize)

	if err != nil {
		panic(err)
	}

	char.Inventory = GetInventoryFromCharID(char.ID)

	char.Pos.X = nx.Maps[char.CurrentMap].Portals[char.CurrentMapPos].X
	char.Pos.Y = nx.Maps[char.CurrentMap].Portals[char.CurrentMapPos].Y

	return char
}
