package def

import (
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/nx"
)

type Character struct {
	ID        int32
	AccountID int32
	WorldID   byte

	MapID       int32
	MapPos      byte
	PreviousMap int32
	PortalCount byte

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
	Guild string

	EquipSlotSize byte
	UseSlotSize   byte
	SetupSlotSize byte
	EtcSlotSize   byte
	CashSlotSize  byte

	Inventory
	Mesos int32

	Skills map[int32]Skill

	MiniGameWins, MiniGameDraw, MiniGameLoss int32
}

func (c Character) Save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mesos=? WHERE id=?`

	// need to calculate nearest spawn point for mapPos

	_, err := database.Handle.Exec(query,
		c.Skin, c.Hair, c.Face, c.Level, c.Job, c.Str, c.Dex, c.Int, c.Luk, c.HP, c.MaxHP, c.MP,
		c.MaxMP, c.AP, c.SP, c.EXP, c.Fame, c.MapID, c.Mesos, c.ID)

	c.Inventory.Save(c.ID)

	// There has to be a better way of doing this in mysql
	for skillID, skill := range c.Skills {
		query = `UPDATE skills SET level=?, cooldown=? WHERE skillID=? AND characterID=?`
		result, err := database.Handle.Exec(query, skill.Level, skill.Cooldown, skillID, c.ID)

		if rows, _ := result.RowsAffected(); rows < 1 || err != nil {
			query = `INSERT INTO skills (characterID, skillID, level, cooldown) VALUES (?, ?, ?, ?)`
			_, err = database.Handle.Exec(query, c.ID, skillID, skill.Level, 0)
		}
	}

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
			&char.MP, &char.MaxMP, &char.AP, &char.SP, &char.EXP, &char.Fame, &char.MapID, &char.MapPos,
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
		&char.MaxMP, &char.AP, &char.SP, &char.EXP, &char.Fame, &char.MapID, &char.MapPos,
		&char.PreviousMap, &char.Mesos, &char.EquipSlotSize, &char.UseSlotSize, &char.SetupSlotSize,
		&char.EtcSlotSize, &char.CashSlotSize)

	if err != nil {
		panic(err)
	}

	char.Inventory = GetInventoryFromCharID(char.ID)

	char.Skills = make(map[int32]Skill)

	for _, s := range GetSkillsFromCharID(char.ID) {
		char.Skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(char.MapID)

	if err != nil {
		panic(err)
	}

	char.Pos.X = nxMap.Portals[char.MapPos].X
	char.Pos.Y = nxMap.Portals[char.MapPos].Y

	return char
}
