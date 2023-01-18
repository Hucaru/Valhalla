package db

import (
	"database/sql"
	"errors"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	"log"
	"time"
)
import _ "github.com/go-sql-driver/mysql"

// DB object used for queries
var Maria *sql.DB

// ConnectToDB - connect to a MySQL instance
func ConnectToDB(user, password, address, port, database string) error {
	var err error
	Maria, err = sql.Open("mysql", user+":"+password+"@tcp("+address+":"+port+")/"+database)
	if err != nil {
		return err
	}
	err = Maria.Ping()

	if err != nil {
		return err
	}

	Maria.SetMaxIdleConns(100)

	return nil
}

func GetLoggedData(uId int64) (model.Player, error) {

	plr := model.Player{}
	Character := model.Character{}

	plr.SetCharacter(Character)
	plr.SetInteraction(model.NewInteraction())

	ch := plr.GetCharacter_P()

	err := Maria.QueryRow(
		"SELECT a.accountID, a.accountID, c.id as characterID, c.channelID, c.nickname, c.hair, c.top, c.bottom, c.clothes, IFNULL(m.time, 0) as time, IFNULL(m.pos_x, 0) as pos_x, IFNULL(m.pos_y, 0) as pos_y, IFNULL(m.pos_z, 0) as pos_z, IFNULL(m.rot_x, 0) as rot_x, IFNULL(m.rot_y, 0) as rot_y, IFNULL(m.rot_z, 0) as rot_z FROM accounts a LEFT JOIN characters c ON c.accountID = a.accountID LEFT JOIN (select characterID, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z, time from movement) as m ON m.characterID = c.id WHERE a.accountID=?  ORDER BY m.time DESC limit 1;", uId).
		Scan(&plr.AccountID,
			&plr.UId, &plr.CharacterID, &plr.RegionID,
			&ch.NickName, &ch.Hair, &ch.Top, &ch.Bottom, &ch.Clothes,
			&ch.Time,
			&ch.PosX, &ch.PosY, &ch.PosZ,
			&ch.RotX, &ch.RotY, &ch.RotZ)

	return plr, err
}

func GetLoggedDataForBot(uuid int64) (*model.Player, error) {

	plr := &model.Player{
		UId:         uuid,
		AccountID:   constant.UNKNOWN,
		CharacterID: constant.UNKNOWN,
		RegionID:    constant.World,
	}

	Character := model.Character{
		Role:     constant.User,
		NickName: "",
		Hair:     "",
		Top:      "",
		Bottom:   "",
		Clothes:  "",
		Time:     constant.DEFAULT_TIME,
		PosX:     constant.PosX,
		PosY:     constant.PosY,
		PosZ:     constant.PosZ,
		RotX:     constant.RotX,
		RotY:     constant.RotY,
		RotZ:     constant.RotZ,
	}

	plr.SetCharacter(Character)
	plr.SetInteraction(model.NewInteraction())

	return plr, nil
}

func GetPlayerAccountIDByNickName(nickname string) (int64, error) {
	accountID := int64(0)
	err := Maria.QueryRow(
		"SELECT a.accountID "+
			"FROM accounts a "+
			"WHERE a.username=? ", nickname).Scan(&accountID)

	return accountID, err
}

func AddNewAccount(msg mc_metadata.C2P_RequestLoginUser) (model.Player, error) {
	pInfo := msg.GetPlayerInfo()

	res, err := Maria.Exec("INSERT INTO accounts (username, password, pin, dob, isLogedIn) VALUES ( ?, ?, ?, ?, ?)",
		pInfo.Nickname, "password", "1", 1, 1)

	if err != nil {
		log.Println("INSERT account", err)
		return model.Player{}, err
	}

	resultPlayer := model.Player{}

	resultPlayer.AccountID, err = res.LastInsertId()
	cRes, cErr := Maria.Exec("INSERT INTO characters "+
		"(accountID, worldID, nickname, "+
		"gender, hair, top, bottom, clothes, channelID) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		resultPlayer.AccountID, 1,
		pInfo.Nickname, 1,
		pInfo.Hair, pInfo.Top, pInfo.Bottom, pInfo.Clothes, constant.World)

	if cErr != nil {
		log.Println("INSERTING ERROR", cErr)
		return model.Player{}, cErr
	}

	spawnPosX := float32(constant.PosX)
	if msg.SpawnPosX != 0 {
		spawnPosX = msg.SpawnPosX
	}

	spawnPosY := float32(constant.PosY)
	if msg.SpawnPosY != 0 {
		spawnPosY = msg.SpawnPosY
	}

	spawnPosZ := float32(constant.PosZ)
	if msg.SpawnPosZ != 0 {
		spawnPosZ = msg.SpawnPosZ
	}

	resultPlayer.CharacterID, err = cRes.LastInsertId()
	AddMovement(resultPlayer.CharacterID,
		spawnPosX, spawnPosY, spawnPosZ,
		constant.RotX, constant.RotY, constant.RotZ)

	ch := model.Character{}
	ch.PosX = spawnPosX
	ch.PosY = spawnPosY
	ch.PosZ = spawnPosZ

	ch.RotX = constant.RotX
	ch.RotY = constant.RotY
	ch.RotZ = constant.RotZ

	ch.NickName = pInfo.Nickname
	ch.Hair = pInfo.Hair
	ch.Top = pInfo.Top
	ch.Bottom = pInfo.Bottom
	ch.Clothes = pInfo.Clothes

	resultPlayer.UId = resultPlayer.AccountID
	resultPlayer.RegionID = constant.World

	resultPlayer.SetCharacter(ch)
	resultPlayer.SetInteraction(model.NewInteraction())

	return resultPlayer, nil
}

func UpdateMovement(
	cID int64,
	posX float32,
	posY float32,
	posZ float32,
	rotX float32,
	rotY float32,
	rotZ float32) error {

	if cID < 0 {
		return errors.New("characterId not found")
	}
	return AddMovement(cID, posX, posY, posZ, rotX, rotY, rotZ)
}

func UpdatePlayerInfo(
	cID int64,
	hair string,
	top string,
	bottom string,
	clothes string) error {
	return updatePlayerInfo(cID, hair, top, bottom, clothes)
}

func updatePlayerInfo(
	cID int64,
	hair string,
	top string,
	bottom string,
	clothes string) error {
	_, err := Maria.Exec("UPDATE characters SET hair=?, top=?, bottom=?, clothes=? WHERE id=?",
		hair, top, bottom, clothes, cID)

	if err != nil {
		log.Println("UPDATING PLAYER INFO ERROR", err)
	}
	return err
}

func AddTranslate(
	originalID int64,
	lng string,
	message string) (int64, error) {
	res, err := Maria.Exec("INSERT INTO message_translates "+
		"(originalID, lng, message) "+
		"VALUES (?, ?, ?)",
		originalID, lng, message)

	if err != nil {
		log.Println("INSERTING ERROR", err)
		return -1, err
	}

	return res.LastInsertId()
}

func FindOriginIDTranslate(message string) (int64, error) {
	var id int64 = -1
	var originalID int64 = -1

	err := Maria.QueryRow("SELECT id, originalID FROM message_translates WHERE message=?", message).Scan(&id, &originalID)

	if err != nil {
		log.Println("FindTranslate SELECT ERROR", err)
		return -1, err
	}

	if originalID > 0 {
		return originalID, nil
	}

	return id, nil
}

func GetTranslate(originalID int64, lng string) (*mc_metadata.P2C_Translate, error) {
	translate := &mc_metadata.P2C_Translate{}

	err := Maria.QueryRow("SELECT lng, message FROM message_translates "+
		"WHERE lng=? AND originalID=?", lng, originalID).Scan(&translate.Code, &translate.Text)

	if err != nil {
		return nil, err
	}

	if len(translate.Code) == 0 {
		return nil, errors.New("not found")
	}

	return translate, nil
}

func AddMovement(
	characterID int64,
	posX float32,
	posY float32,
	posZ float32,
	rotX float32,
	rotY float32,
	rotZ float32) error {
	_, err := Maria.Exec("INSERT INTO movement "+
		"(characterID, pos_x, pos_y, pos_z, rot_x, rot_y, rot_z, time) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		characterID, posX, posY, posZ, rotX, rotY, rotZ, time.Now().UnixNano()/int64(time.Millisecond))

	if err != nil {
		log.Println("INSERTING ERROR", err)
	}
	return err
}

func AddPublicMessage(cID int64, regionID int64, text string) {
	addChatMessage(cID, regionID, text, constant.NO_TARGET)
}

func AddWhisperMessage(cID int64, targetCID int64, text string) {
	addChatMessage(cID, constant.World, text, targetCID)
}

func addChatMessage(
	characterID int64,
	regionID int64,
	text string,
	targetID int64) {
	_, err := Maria.Exec("INSERT INTO chat "+
		"(characterID, regionID, text, targetID, createdAt) "+
		"VALUES (?, ?, ?, ?, ?)",
		characterID, regionID, text, targetID, time.Now().UnixNano()/int64(time.Millisecond))

	if err != nil {
		log.Println("INSERTING ERROR", err)
	}
}

func UpdateLoginState(uUID int64, isLogedIn bool) error {
	in := 0
	if isLogedIn {
		in = 1
	} else {
		in = 0
	}
	_, err := Maria.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", in, uUID)
	return err
}

func UpdateRegionID(cID int64, channelID int32) error {
	_, err := Maria.Exec("UPDATE characters SET channelID=? WHERE accountID=?", channelID, cID)
	return err
}

func ResetLoginState(isLogedIn bool) error {
	in := 0
	if isLogedIn {
		in = 1
	} else {
		in = 0
	}
	_, err := Maria.Exec("UPDATE accounts SET isLogedIn=?", in)
	return err
}
