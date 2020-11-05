package player

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/db"
	"github.com/Hucaru/Valhalla/server/item"
)

// GetCharactersFromAccountWorldID - characters under a specific account
func GetCharactersFromAccountWorldID(accountID int32, worldID byte) []Data {
	c := []Data{}

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize"

	chars, err := db.DB.Query("SELECT "+filter+" FROM characters WHERE accountID=? AND worldID=?", accountID, worldID)

	if err != nil {
		log.Println(err)
	}

	defer chars.Close()

	for chars.Next() {
		var char Data

		err = chars.Scan(&char.id, &char.accountID, &char.worldID, &char.name, &char.gender, &char.skin, &char.hair,
			&char.face, &char.level, &char.job, &char.str, &char.dex, &char.intt, &char.luk, &char.hp, &char.maxHP,
			&char.mp, &char.maxMP, &char.ap, &char.sp, &char.exp, &char.fame, &char.mapID, &char.mapPos,
			&char.previousMap, &char.mesos, &char.equipSlotSize, &char.useSlotSize, &char.setupSlotSize,
			&char.etcSlotSize, &char.cashSlotSize)

		if err != nil {
			log.Println(err)
		}

		char.equip, char.use, char.setUp, char.etc, char.cash = item.LoadInventoryFromDb(char.id)

		c = append(c, char)
	}

	return c
}

// LoadFromID - player id to load from database
func LoadFromID(id int32, conn mnet.Client) Data {
	c := Data{}
	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize,miniGameWins," +
		"miniGameDraw,miniGameLoss,miniGamePoints"

	err := db.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&c.id,
		&c.accountID, &c.worldID, &c.name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos,
		&c.previousMap, &c.mesos, &c.equipSlotSize, &c.useSlotSize, &c.setupSlotSize,
		&c.etcSlotSize, &c.cashSlotSize, &c.miniGameWins, &c.miniGameDraw, &c.miniGameLoss,
		&c.miniGamePoints)

	if err != nil {
		panic(err)
	}

	c.skills = make(map[int32]Skill)

	for _, s := range getSkillsFromCharID(c.id) {
		c.skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		log.Println(err)
	}

	c.pos.SetX(nxMap.Portals[c.mapPos].X)
	c.pos.SetY(nxMap.Portals[c.mapPos].Y)

	c.equip, c.use, c.setUp, c.etc, c.cash = item.LoadInventoryFromDb(c.id)
	c.conn = conn
	return c
}
