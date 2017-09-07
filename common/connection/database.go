package connection

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func ConnectToDb() {
	var err error
	Db, err = sql.Open("mysql", "root:password@/maplestory")

	if err != nil {
		log.Fatal(err.Error())
	}

	err = Db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}
}
