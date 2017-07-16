package connection

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func ConnectToDb() {
	var err error
	Db, err = sql.Open("mysql", "root:password@/maplestory")

	if err != nil {
		panic(err.Error())
	}

	err = Db.Ping()

	if err != nil {
		panic(err.Error())
	}
}
