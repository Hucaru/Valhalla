package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
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

	Maria.SetMaxIdleConns(10)

	return nil
}

func GetLoggedData(uUID string) (model.Account, error) {

	acc := &model.Account{
		UId:         uUID,
		AccountID:   -1,
		CharacterID: -1,
		RegionID:    constant.All,
		Hair:        "",
		Top:         "",
		Bottom:      "",
		Clothes:     "",
		Time:        0,
		PosX:        constant.PosX,
		PosY:        constant.PosY,
		PosZ:        constant.PosZ,
		RotX:        constant.RotX,
		RotY:        constant.RotY,
		RotZ:        constant.RotZ,
	}

	err := Maria.QueryRow(
		"SELECT a.accountID, a.uId, c.id as characterID, c.channelID, "+
			"c.hair, c.top, c.bottom, c.clothes, "+
			"IFNULL(m.time, 0) as time, "+
			"IFNULL(m.pos_x, 0) as pos_x, "+
			"IFNULL(m.pos_y, 0) as pos_y, "+
			"IFNULL(m.pos_z, 0) as pos_z, "+
			"IFNULL(m.rot_x, 0) as rot_x, "+
			"IFNULL(m.rot_y, 0) as rot_y, "+
			"IFNULL(m.rot_z, 0) as rot_z "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"LEFT JOIN movement m ON m.characterID = characterID "+
			"WHERE a.uId=? "+
			"ORDER BY time DESC "+
			"LIMIT 1", uUID).
		Scan(&acc.AccountID,
			&acc.UId,
			&acc.CharacterID,
			&acc.RegionID,
			&acc.Hair, &acc.Top, &acc.Bottom, &acc.Clothes,
			&acc.Time, &acc.PosX, &acc.PosY, &acc.PosZ, &acc.RotX, &acc.RotY, &acc.RotZ)

	return *acc, err
}

func GetLoggedDataByName(uUID string, nickname string) (model.Account, error) {

	acc := &model.Account{
		UId:         uUID,
		AccountID:   -1,
		CharacterID: -1,
		RegionID:    constant.All,
		Hair:        "",
		Top:         "",
		Bottom:      "",
		Clothes:     "",
		Time:        0,
		PosX:        constant.PosX,
		PosY:        constant.PosY,
		PosZ:        constant.PosZ,
		RotX:        constant.RotX,
		RotY:        constant.RotY,
		RotZ:        constant.RotZ,
	}

	err := Maria.QueryRow(
		"SELECT a.accountID, a.uId, c.id as characterID, c.channelID, "+
			"c.hair, c.top, c.bottom, c.clothes, "+
			"IFNULL(m.time, 0) as time, "+
			"IFNULL(m.pos_x, 0) as pos_x, "+
			"IFNULL(m.pos_y, 0) as pos_y, "+
			"IFNULL(m.pos_z, 0) as pos_z, "+
			"IFNULL(m.rot_x, 0) as rot_x, "+
			"IFNULL(m.rot_y, 0) as rot_y, "+
			"IFNULL(m.rot_z, 0) as rot_z "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"LEFT JOIN movement m ON m.characterID = characterID "+
			"WHERE c.nickname=? "+
			"ORDER BY time DESC "+
			"LIMIT 1", nickname).
		Scan(&acc.AccountID,
			&acc.UId,
			&acc.CharacterID,
			&acc.RegionID,
			&acc.Hair, &acc.Top, &acc.Bottom, &acc.Clothes,
			&acc.Time, &acc.PosX, &acc.PosY, &acc.PosZ, &acc.RotX, &acc.RotY, &acc.RotZ)

	return *acc, err
}

func GetLoggedUsersData(uUID string, regionID int64) ([]*model.Account, error) {

	accounts := make([]*model.Account, 0)

	rows, err := Maria.Query(
		"SELECT a.accountID, a.uId, c.id as characterID, a.isLogedIn, c.channelID, "+
			"c.hair, c.top, c.bottom, c.clothes, "+
			"IFNULL(m.time, 0) as time, "+
			"IFNULL(m.pos_x, 0) as pos_x, "+
			"IFNULL(m.pos_y, 0) as pos_y, "+
			"IFNULL(m.pos_z, 0) as pos_z, "+
			"IFNULL(m.rot_x, 0) as rot_x, "+
			"IFNULL(m.rot_y, 0) as rot_y, "+
			"IFNULL(m.rot_z, 0) as rot_z "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"LEFT JOIN (SELECT * FROM movement m ORDER BY m.time DESC LIMIT 1) m ON m.characterID = c.id  "+
			"WHERE a.uId != ? AND a.isLogedIn != 0 AND c.channelID = ? "+
			"ORDER BY time DESC", uUID, regionID)

	if err != nil {
		log.Println("LOGGED USERS SELECTING ERROR", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var is int32
		acc := &model.Account{
			UId:         uUID,
			AccountID:   -1,
			RegionID:    constant.All,
			CharacterID: -1,
			Time:        0,
			Hair:        "",
			Top:         "",
			Bottom:      "",
			Clothes:     "",
			PosX:        constant.PosX,
			PosY:        constant.PosY,
			PosZ:        constant.PosZ,
			RotX:        constant.RotX,
			RotY:        constant.RotY,
			RotZ:        constant.RotZ,
		}

		if err := rows.Scan(
			&acc.AccountID, &acc.UId, &acc.CharacterID, &is, &acc.RegionID,
			&acc.Hair, &acc.Top, &acc.Bottom, &acc.Clothes,
			&acc.Time,
			&acc.PosX, &acc.PosY, &acc.PosZ, &acc.RotX, &acc.RotY, &acc.RotZ); err != nil {
			log.Println("LOGGED USERS SELECTING ERROR", err)
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		log.Println("LOGGED USERS SELECTING ERROR", err)
		return nil, err
	}

	return accounts, nil
}

func InsertNewAccount(uUid string, conn mnet.Client) error {
	res, err := Maria.Exec("INSERT INTO accounts (uId, username, password, pin, dob, isLogedIn) VALUES (?, ?, ?, ?, ?, ?)",
		uUid, "test", "password", "1", 1, 1)

	if err != nil {
		log.Println(err)
		return err
	}

	accountID, err := res.LastInsertId()
	conn.SetAccountID(int32(accountID))
	cRes, cErr := Maria.Exec("INSERT INTO characters "+
		"(accountID, worldID, nickname, gender, skin, hair, face, str, dex, intt, luk, top, bottom, clothes, channelID) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		accountID, 1, fmt.Sprintf("player#%d", time.Now().UnixNano()/int64(time.Millisecond)), 1, 1, "", 1, 1, 1, 1, 1, "", "", "", constant.All)

	if cErr != nil {
		log.Println("INSERTING ERROR", cErr)
		return cErr
	}

	characterID, err := cRes.LastInsertId()
	return insertMovement(characterID, constant.PosX, constant.PosY, constant.PosZ, constant.RotX, constant.RotY, constant.RotZ)
}

func UpdateMovement(
	uID string,
	posX float32,
	posY float32,
	posZ float32,
	rotX float32,
	rotY float32,
	rotZ float32) error {

	cID := findCharacterByUid(uID)
	if cID < 0 {
		return errors.New("characterId not found")
	}
	return insertMovement(cID, posX, posY, posZ, rotX, rotY, rotZ)
}

func UpdatePlayerInfo(
	uID string,
	nickname string,
	hair string,
	top string,
	bottom string,
	clothes string) error {

	cID := findCharacterByUid(uID)
	if cID < 0 {
		return insertPlayerInfo(nickname, hair, top, bottom, clothes)
	}
	return updatePlayerInfo(cID, hair, top, bottom, clothes)
}

func findCharacterByUid(uID string) int64 {
	var accountID int64
	var characterID int64

	err := Maria.QueryRow(
		"SELECT a.accountID, c.id as characterID "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"WHERE a.uId=? "+
			"LIMIT 1", uID).
		Scan(&accountID, &characterID)
	if err != nil {
		return -1
	}
	return characterID
}

func insertPlayerInfo(
	nickname string,
	hair string,
	top string,
	bottom string,
	clothes string) error {
	_, err := Maria.Exec("INSERT INTO characters "+
		"(nickname, hair, top, bottom, clothes) "+
		"VALUES (?, ?, ?, ?, ?)",
		nickname, hair, top, bottom, clothes)

	if err != nil {
		log.Println("INSERTING PLAYER INFO ERROR", err)
	}
	return err
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

func insertMovement(
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

func InsertPublicMessage(uID string, regionID int64, text string) {
	var accountID int64
	var characterID int64

	err := Maria.QueryRow(
		"SELECT a.accountID, c.id as characterID "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"WHERE a.uId=? "+
			"LIMIT 1", uID).
		Scan(&accountID, &characterID)
	if err != nil {
		log.Println("ERROR SELECTING ACCOUNT")
	}

	insertChatMessage(characterID, regionID, text, constant.NO_TARGET)
}

func InsertWhisperMessage(uID string, targetID string, text string) string {
	var accountID int64
	var characterID int64
	var targetUID string
	var targetCID int64

	err := Maria.QueryRow(
		"SELECT a1.accountID, c1.id as characterID, a2.uId as targetUID, IFNULL(c2.id, -1) as TargetCID "+
			"FROM accounts a1 "+
			"LEFT JOIN characters c1 ON c1.accountID = a1.accountID "+
			"LEFT JOIN characters c2 ON c2.nickname = ? "+
			"LEFT JOIN accounts a2 ON a2.accountID = c2.accountID "+
			"WHERE a1.uId=? "+
			"LIMIT 1", targetID, uID).
		Scan(&accountID, &characterID, &targetUID, &targetCID)
	if err != nil {
		log.Println("ERROR SELECTING ACCOUNT")
	}

	insertChatMessage(characterID, constant.All, text, targetCID)
	return targetUID
}

func insertChatMessage(
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

func UpdateLoginState(uUID string, isLogedIn bool) error {
	in := 0
	if isLogedIn {
		in = 1
	} else {
		in = 0
	}
	_, err := Maria.Exec("UPDATE accounts SET isLogedIn=? WHERE uId=?", in, uUID)
	return err
}

func UpdateRegionID(uid string, channelID int32) error {
	cID := findCharacterByUid(uid)
	_, err := Maria.Exec("UPDATE characters SET channelID=? WHERE id=?", channelID, cID)
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
