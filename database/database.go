package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func Connect(user, password, address, port, database string) {
	var err error
	Db, err = sql.Open("mysql", user+":"+password+"@tcp("+address+":"+port+")/"+database)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer Db.Close()

	err = Db.Ping()

	if err != nil {
		log.Fatal(err.Error()) // change to attempt to re-connect
	}

	log.Println("Connected to database")

	timer := time.NewTicker(60 * time.Second) // short enough ping time?

	for {
		select {
		case <-timer.C:
			err = Db.Ping()

			if err != nil {
				log.Fatal(err.Error()) // change to attempt to re-connect
			}
		}
	}
}
