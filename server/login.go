package server

import (
	"database/sql"
	"log"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// LoginServer state
type LoginServer struct {
	migrating map[mnet.Client]bool
	db        *sql.DB
	worlds    []world
}

// Initialise the server
func (server *LoginServer) Initialise(dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.migrating = make(map[mnet.Client]bool)

	var err error
	server.db, err = sql.Open("mysql", dbuser+":"+dbpassword+"@tcp("+dbaddress+":"+dbport+")/"+dbdatabase)

	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.db.Ping()

	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Connected to database")

	server.CleanupDB()

	log.Println("Cleaned up the database")
}

// CleanupDB sets all accounts isLogedIn to 0
func (server *LoginServer) CleanupDB() {
	res, err := server.db.Exec("UPDATE accounts AS a INNER JOIN characters c ON a.accountID = c.accountID SET a.isLogedIn = 0 WHERE isLogedIn = 1 AND a.accountID != ALL (SELECT c.accountID FROM characters c WHERE c.channelID != -1);")
	if err != nil {
		log.Fatal(err)
	}
	amount, _ := res.RowsAffected()
	log.Printf("Set %d isLogedin rows to 0.", amount)
}

// HandleServerPacket from world
func (server *LoginServer) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldNew:
		server.handleNewWorld(conn, reader)
	case opcode.WorldInfo:
		server.handleWorldInfo(conn, reader)
	default:
		log.Println("UNKNOWN WORLD PACKET:", reader)
	}
}

// ServerDisconnected handler
func (server *LoginServer) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.worlds {
		if v.conn == conn {
			log.Println(v.name, "disconnected")
			server.worlds[i].conn = nil
			server.worlds[i].channels = []channel{}
			break
		}
	}
}

// The following logic could do with being cleaned up
func (server *LoginServer) handleNewWorld(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Server register request from", conn)
	if len(server.worlds) > 14 {
		log.Println("Rejected")
		conn.Send(mpacket.CreateInternal(opcode.WorldRequestBad))
	} else {
		name := reader.ReadString(reader.ReadInt16())

		if name == "" {
			name = constant.WORLD_NAMES[len(server.worlds)]

			registered := false
			for i, v := range server.worlds {
				if v.conn == nil {
					server.worlds[i].conn = conn
					name = server.worlds[i].name

					registered = true
					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, world{conn: conn, name: name})
			}

			p := mpacket.CreateInternal(opcode.WorldRequestOk)
			p.WriteString(name)
			conn.Send(p)

			log.Println("Registered", name)
		} else {
			registered := false
			for i, w := range server.worlds {
				if w.name == name {
					server.worlds[i].conn = conn
					server.worlds[i].name = name

					p := mpacket.CreateInternal(opcode.WorldRequestOk)
					p.WriteString(name)
					conn.Send(p)

					registered = true

					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, world{conn: conn, name: name})

				p := mpacket.CreateInternal(opcode.WorldRequestOk)
				p.WriteString(server.worlds[len(server.worlds)-1].name)
				conn.Send(p)
			}

			log.Println("Re-registered", name)
		}
	}
}

func (server *LoginServer) handleWorldInfo(conn mnet.Server, reader mpacket.Reader) {
	for i, v := range server.worlds {
		if v.conn != conn {
			continue
		}

		server.worlds[i].serialisePacket(reader)

		if v.name == "" {
			log.Println("Registerd new world", server.worlds[i].name)
		} else {
			log.Println("Updated world info for", v.name)
		}
	}
}

// ClientDisconnected from server
func (server *LoginServer) ClientDisconnected(conn mnet.Client) {
	if isMigrating, ok := server.migrating[conn]; ok && isMigrating {
		delete(server.migrating, conn)
	} else {
		_, err := server.db.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

		if err != nil {
			log.Println("Unable to complete logout for ", conn.GetAccountID())
		}
	}

	conn.Cleanup()
}
