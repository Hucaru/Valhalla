package game

import (
	"database/sql"

	"github.com/Hucaru/Valhalla/nx"
)

type character struct {
	id        int32
	accountID int32
	worldID   byte

	mapID       int32
	mapPos      byte
	previousMap int32
	portalCount byte

	job int16

	level byte
	str   int16
	dex   int16
	intt  int16
	luk   int16
	hp    int16
	maxHP int16
	mp    int16
	maxMP int16
	ap    int16
	sp    int16
	exp   int32
	fame  int16

	name     string
	gender   byte
	skin     byte
	face     int32
	hair     int32
	chairID  int32
	stance   byte
	pos      pos
	foothold int16
	guild    string

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	inventory
	mesos int32

	Skills map[int32]Skill

	MinigameWins, MinigameDraw, MinigameLoss int32
}

func (c character) save(db *sql.DB) error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mesos=? WHERE id=?`

	// need to calculate nearest spawn point for mapPos

	_, err := db.Exec(query,
		c.skin, c.hair, c.face, c.level, c.job, c.str, c.dex, c.intt, c.luk, c.hp, c.maxHP, c.mp,
		c.maxMP, c.ap, c.sp, c.exp, c.fame, c.mapID, c.mesos, c.id)

	c.inventory.save(c.id)

	// There has to be a better way of doing this in mysql
	for skillID, skill := range c.Skills {
		query = `UPDATE skills SET level=?, cooldown=? WHERE skillID=? AND characterID=?`
		result, err := db.Exec(query, skill.Level, skill.Cooldown, skillID, c.id)

		if rows, _ := result.RowsAffected(); rows < 1 || err != nil {
			query = `INSERT INTO skills (characterID, skillID, level, cooldown) VALUES (?, ?, ?, ?)`
			_, err = db.Exec(query, c.id, skillID, skill.Level, 0)
		}
	}

	return err
}

func getCharactersFromAccountWorldID(db *sql.DB, accountID int32, worldID byte) []character {
	c := []character{}

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	chars, err := db.Query("SELECT "+filter+" FROM characters WHERE accountID=? AND worldID=?", accountID, worldID)

	if err != nil {
		panic(err)
	}

	defer chars.Close()

	for chars.Next() {
		var char character

		err = chars.Scan(&char.id, &char.accountID, &char.worldID, &char.name, &char.gender, &char.skin, &char.hair,
			&char.face, &char.level, &char.job, &char.str, &char.dex, &char.intt, &char.luk, &char.hp, &char.maxHP,
			&char.mp, &char.maxMP, &char.ap, &char.sp, &char.exp, &char.fame, &char.mapID, &char.mapPos,
			&char.previousMap, &char.mesos, &char.equipSlotSize, &char.useSlotSize, &char.setupSlotSize,
			&char.etcSlotSize, &char.cashSlotSize)

		if err != nil {
			panic(err)
		}

		char.inventory = getInventoryFromCharID(db, char.id)

		c = append(c, char)
	}

	return c
}

func (c *character) loadFromID(db *sql.DB, id int32) {
	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	err := db.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&c.id,
		&c.accountID, &c.worldID, &c.name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos,
		&c.previousMap, &c.mesos, &c.equipSlotSize, &c.useSlotSize, &c.setupSlotSize,
		&c.etcSlotSize, &c.cashSlotSize)

	if err != nil {
		panic(err)
	}

	c.inventory = getInventoryFromCharID(db, c.id)

	c.Skills = make(map[int32]Skill)

	for _, s := range GetSkillsFromCharID(db, c.id) {
		c.Skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		panic(err)
	}

	c.pos.x = nxMap.Portals[c.mapPos].X
	c.pos.y = nxMap.Portals[c.mapPos].Y
}
