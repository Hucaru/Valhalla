package db

import (
	"database/sql"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/mnet"
	"log"
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
		Time:        0,
		PosX:        0,
		PosY:        0,
		PosZ:        0,
		RotX:        0,
		RotY:        0,
		RotZ:        0,
	}

	err := Maria.QueryRow(
		"SELECT a.accountID, a.u_id, c.id as characterID, "+
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
			"WHERE a.u_id=? "+
			"ORDER BY time DESC "+
			"LIMIT 1", uUID).
		Scan(&acc.AccountID, &acc.UId, &acc.CharacterID, &acc.Time, &acc.PosX, &acc.PosY, &acc.PosZ, &acc.RotX, &acc.RotY, &acc.RotZ)

	return *acc, err
}

func InsertNewAccount(uUid string, conn mnet.Client) {
	res, err := Maria.Exec("INSERT INTO accounts (u_id, username, password, pin, dob, isLogedIn) VALUES (?, ?, ?, ?, ?, ?)",
		uUid, "test", "password", "1", 1, 1)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	accountID, err := res.LastInsertId()
	conn.SetAccountID(int32(accountID))

	res, err = Maria.Exec("INSERT INTO characters "+
		"(accountID, worldID, name, gender, skin, hair, face, str, dex, intt, luk) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		accountID, 1, "test", 1, 1, 1, 1, 1, 1, 1, 1)

	if err != nil {
		log.Println("INSERTING ERROR", err)
		panic(err)
	}
}

func UpdateLoginState(uUID string, isLogedIn bool) error {
	in := 0
	if isLogedIn {
		in = 1
	} else {
		in = 0
	}
	_, err := Maria.Exec("UPDATE accounts SET isLogedIn=? WHERE u_id=?", in, uUID)
	return err
}
