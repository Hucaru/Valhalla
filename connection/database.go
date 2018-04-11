package connection

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func ConnectToDb() {
	var err error
	Db, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+
		os.Getenv("DB_ADDRESS")+":"+os.Getenv("DB_PORT")+")/"+os.Getenv("DB_DATABASE"))

	if err != nil {
		log.Fatal(err.Error())
	}

	err = Db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}
}
