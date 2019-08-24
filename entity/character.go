package entity

import (
	"database/sql"

	"github.com/Hucaru/Valhalla/nx"
)

type Character struct {
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

	inventory Inventory
	mesos     int32

	skills map[int32]Skill

	minigameWins, minigameDraw, minigameLoss int32
}

func (c Character) ID() int32               { return c.id }
func (c Character) AccountID() int32        { return c.accountID }
func (c Character) WorldID() byte           { return c.worldID }
func (c Character) MapID() int32            { return c.mapID }
func (c Character) MapPos() byte            { return c.mapPos }
func (c Character) PreviousMap() int32      { return c.previousMap }
func (c Character) PortalCount() byte       { return c.portalCount }
func (c Character) Job() int16              { return c.job }
func (c Character) Level() byte             { return c.level }
func (c Character) Str() int16              { return c.str }
func (c Character) Dex() int16              { return c.dex }
func (c Character) Int() int16              { return c.intt }
func (c Character) Luk() int16              { return c.luk }
func (c Character) HP() int16               { return c.hp }
func (c Character) MaxHP() int16            { return c.maxHP }
func (c Character) MP() int16               { return c.mp }
func (c Character) MaxMP() int16            { return c.maxMP }
func (c Character) AP() int16               { return c.ap }
func (c Character) SP() int16               { return c.sp }
func (c Character) Exp() int32              { return c.exp }
func (c Character) Fame() int16             { return c.fame }
func (c Character) Name() string            { return c.name }
func (c Character) Gender() byte            { return c.gender }
func (c Character) Skin() byte              { return c.skin }
func (c Character) Face() int32             { return c.face }
func (c Character) Hair() int32             { return c.hair }
func (c Character) ChairID() int32          { return c.chairID }
func (c Character) Stance() byte            { return c.stance }
func (c Character) Pos() pos                { return c.pos }
func (c Character) Foothold() int16         { return c.foothold }
func (c Character) Guild() string           { return c.guild }
func (c Character) EquipSlotSize() byte     { return c.equipSlotSize }
func (c Character) UseSlotSize() byte       { return c.useSlotSize }
func (c Character) SetupSlotSize() byte     { return c.setupSlotSize }
func (c Character) EtcSlotSize() byte       { return c.etcSlotSize }
func (c Character) CashSlotSize() byte      { return c.cashSlotSize }
func (c Character) Inventory() Inventory    { return c.inventory }
func (c Character) Mesos() int32            { return c.mesos }
func (c Character) Skills() map[int32]Skill { return c.skills }
func (c Character) MinigameWins() int32     { return c.minigameWins }
func (c Character) MinigameDraw() int32     { return c.minigameDraw }
func (c Character) MinigameLoss() int32     { return c.minigameLoss }

func (c Character) Save(db *sql.DB, inst instance) error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=? WHERE id=?`

	// need to calculate nearest spawn point for mapPos
	portal, err := inst.CalculateNearestSpawnPortal(c.pos)

	if err == nil {
		c.mapPos = portal.ID()
	}

	_, err = db.Exec(query,
		c.skin, c.hair, c.face, c.level, c.job, c.str, c.dex, c.intt, c.luk, c.hp, c.maxHP, c.mp,
		c.maxMP, c.ap, c.sp, c.exp, c.fame, c.mapID, c.mapPos, c.mesos, c.id)

	c.inventory.save(c.id)

	// There has to be a better way of doing this in mysql
	for skillID, skill := range c.skills {
		query = `UPDATE skills SET level=?, cooldown=? WHERE skillID=? AND characterID=?`
		result, err := db.Exec(query, skill.Level, skill.Cooldown, skillID, c.id)

		if rows, _ := result.RowsAffected(); rows < 1 || err != nil {
			query = `INSERT INTO skills (characterID, skillID, level, cooldown) VALUES (?, ?, ?, ?)`
			_, err = db.Exec(query, c.id, skillID, skill.Level, 0)
		}
	}

	return err
}

func GetCharactersFromAccountWorldID(db *sql.DB, accountID int32, worldID byte) []Character {
	c := []Character{}

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	chars, err := db.Query("SELECT "+filter+" FROM characters WHERE accountID=? AND worldID=?", accountID, worldID)

	if err != nil {
		panic(err)
	}

	defer chars.Close()

	for chars.Next() {
		var char Character

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

func (c *Character) LoadFromID(db *sql.DB, id int32) {
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

	c.skills = make(map[int32]Skill)

	for _, s := range getSkillsFromCharID(db, c.id) {
		c.skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		panic(err)
	}

	c.pos.x = nxMap.Portals[c.mapPos].X
	c.pos.y = nxMap.Portals[c.mapPos].Y
}
