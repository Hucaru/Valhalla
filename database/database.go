package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Handle *sql.DB

func Connect(user, password, address, port, database string) {
	var err error
	Handle, err = sql.Open("mysql", user+":"+password+"@tcp("+address+":"+port+")/"+database)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = Handle.Ping()

	if err != nil {
		log.Fatal(err.Error()) // change to attempt to re-connect
	}

	log.Println("Connected to database")
}

func Monitor() {
	timer := time.NewTicker(60 * time.Second) // short enough ping time?

	for {
		select {
		case <-timer.C:
			err := Handle.Ping()

			if err != nil {
				log.Fatal(err.Error()) // change to attempt to re-connect
			}
		}
	}
}
