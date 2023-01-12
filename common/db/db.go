package db

import (
	"database/sql"
	"errors"
	"fmt"
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

func GetLoggedData(uUID string) (*model.Player, error) {

	plr := &model.Player{
		UId:         uUID,
		AccountID:   constant.UNKNOWN,
		CharacterID: constant.UNKNOWN,
		RegionID:    constant.World,
		Character: &model.Character{
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
		},
		Interaction: nil,
	}

	err := Maria.QueryRow(
		"SELECT a.accountID, "+
			"a.uId, c.id as characterID, c.channelID, "+
			"c.nickname, c.hair, c.top, c.bottom, c.clothes, "+
			"IFNULL(m.time, 0) as time, "+
			"IFNULL(m.pos_x, 0) as pos_x, IFNULL(m.pos_y, 0) as pos_y, IFNULL(m.pos_z, 0) as pos_z, "+
			"IFNULL(m.rot_x, 0) as rot_x, IFNULL(m.rot_y, 0) as rot_y, IFNULL(m.rot_z, 0) as rot_z "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"LEFT JOIN movement m ON m.characterID = characterID "+
			"WHERE a.uId=? "+
			"ORDER BY time DESC "+
			"LIMIT 1", uUID).
		Scan(&plr.AccountID,
			&plr.UId, &plr.CharacterID, &plr.RegionID,
			&plr.Character.NickName, &plr.Character.Hair, &plr.Character.Top, &plr.Character.Bottom, &plr.Character.Clothes,
			&plr.Character.Time,
			&plr.Character.PosX, &plr.Character.PosY, &plr.Character.PosZ,
			&plr.Character.RotX, &plr.Character.RotY, &plr.Character.RotZ)

	return plr, err
}

func GetLoggedDataForBot(uUID string) (*model.Player, error) {

	plr := &model.Player{
		UId:         uUID,
		AccountID:   constant.UNKNOWN,
		CharacterID: constant.UNKNOWN,
		RegionID:    constant.World,
		Character: &model.Character{
			Role:     constant.User,
			NickName: uUID,
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
		},
		Interaction: nil,
	}

	return plr, nil
}

func GetLoggedDataByName(req *mc_metadata.C2P_RequestPlayerInfo) (*model.Player, error) {

	plr := &model.Player{
		UId:         req.GetUuId(),
		AccountID:   constant.UNKNOWN,
		CharacterID: constant.UNKNOWN,
		RegionID:    constant.World,
		Character: &model.Character{
			Role:     constant.User,
			NickName: req.GetNickname(),
			Hair:     req.GetHair(),
			Top:      req.GetTop(),
			Bottom:   req.GetBottom(),
			Clothes:  req.GetClothes(),
			Time:     constant.DEFAULT_TIME,
			PosX:     constant.PosX,
			PosY:     constant.PosY,
			PosZ:     constant.PosZ,
			RotX:     constant.RotX,
			RotY:     constant.RotY,
			RotZ:     constant.RotZ,
		},
		Interaction: nil,
	}

	err := Maria.QueryRow(
		"SELECT a.accountID, "+
			"a.uId, c.id as characterID, c.channelID, "+
			"c.nickname, c.hair, c.top, c.bottom, c.clothes, "+
			"IFNULL(m.time, 0) as time, "+
			"IFNULL(m.pos_x, 0) as pos_x, IFNULL(m.pos_y, 0) as pos_y, IFNULL(m.pos_z, 0) as pos_z, "+
			"IFNULL(m.rot_x, 0) as rot_x, IFNULL(m.rot_y, 0) as rot_y, IFNULL(m.rot_z, 0) as rot_z "+
			"FROM accounts a "+
			"LEFT JOIN characters c ON c.accountID = a.accountID "+
			"LEFT JOIN movement m ON m.characterID = characterID "+
			"WHERE a.nickname=? "+
			"ORDER BY time DESC "+
			"LIMIT 1", req.GetNickname()).
		Scan(&plr.AccountID,
			&plr.UId, &plr.CharacterID, &plr.RegionID,
			&plr.Character.NickName, &plr.Character.Hair, &plr.Character.Top, &plr.Character.Bottom, &plr.Character.Clothes,
			&plr.Character.Time,
			&plr.Character.PosX, &plr.Character.PosY, &plr.Character.PosZ,
			&plr.Character.RotX, &plr.Character.RotY, &plr.Character.RotZ)

	return plr, err
}

func AddNewAccount(plr *model.Player) error {
	res, err := Maria.Exec("INSERT INTO accounts (uId, username, password, pin, dob, isLogedIn) VALUES (?, ?, ?, ?, ?, ?)",
		plr.UId, "test", "password", "1", 1, 1)

	if err != nil {
		log.Println("INSERT account", err)
		return err
	}
	err = nil

	if plr.Character.NickName == "" {
		plr.Character.NickName = fmt.Sprintf("player#%d", time.Now().UnixNano()/int64(time.Millisecond))
	}

	plr.AccountID, err = res.LastInsertId()
	cRes, cErr := Maria.Exec("INSERT INTO characters "+
		"(accountID, worldID, nickname, "+
		"gender, hair, top, bottom, clothes, channelID) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		plr.AccountID, 1,
		plr.Character.NickName, 1,
		plr.Character.Hair, plr.Character.Top, plr.Character.Bottom, plr.Character.Clothes, constant.World)

	if cErr != nil {
		log.Println("INSERTING ERROR", cErr)
		return cErr
	}
	err = nil
	plr.CharacterID, err = cRes.LastInsertId()
	return AddMovement(plr.CharacterID,
		constant.PosX, constant.PosY, constant.PosZ,
		constant.RotX, constant.RotY, constant.RotZ)
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

func UpdateRegionID(cID int64, channelID int32) error {
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
