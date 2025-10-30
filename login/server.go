package login

import (
	"log"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
)

// Server state
type Server struct {
	migrating    map[mnet.Client]bool
	// db        *sql.DB
	worlds       []internal.World
	withPin      bool
	autoRegister bool
}

// Initialise the server
func (server *Server) Initialise(dbuser, dbpassword, dbaddress, dbport, dbdatabase string, withpin bool, autoRegister bool) {
	server.migrating = make(map[mnet.Client]bool)
	server.withPin = withpin
	server.autoRegister = autoRegister

	err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase)

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Connected to database")

	server.CleanupDB()

	log.Println("Cleaned up the database")
}

// CleanupDB sets all accounts isLogedIn to 0
func (server *Server) CleanupDB() {
	res, err := common.DB.Exec("UPDATE accounts AS a INNER JOIN characters c ON a.accountID = c.accountID SET a.isLogedIn = 0 WHERE isLogedIn = 1 AND a.accountID != ALL (SELECT c.accountID FROM characters c WHERE c.channelID != -1);")
	if err != nil {
		log.Fatal(err)
	}
	amount, _ := res.RowsAffected()
	log.Printf("Set %d isLogedin rows to 0.", amount)
}

// ServerDisconnected handler
func (server *Server) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.worlds {
		if v.Conn == conn {
			log.Println(v.Name, "disconnected")
			server.worlds[i].Conn = nil
			server.worlds[i].Channels = []internal.Channel{}
			break
		}
	}
}

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	if isMigrating, ok := server.migrating[conn]; ok && isMigrating {
		delete(server.migrating, conn)
	} else {
		_, err := common.DB.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}

	conn.Cleanup()
}
