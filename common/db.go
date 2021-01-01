package common

import "database/sql"

// DB object used for queries
var DB *sql.DB

// ConnectToDB - connect to a MySQL instance
func ConnectToDB(user, password, address, port, database string) error {
	var err error
	DB, err = sql.Open("mysql", user+":"+password+"@tcp("+address+":"+port+")/"+database)

	if err != nil {
		return err
	}

	err = DB.Ping()

	if err != nil {
		return err
	}

	DB.SetMaxIdleConns(10)

	return nil
}
