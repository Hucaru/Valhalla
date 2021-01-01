package login

import (
	"log"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Server state
type Server struct {
	migrating map[mnet.Client]bool
	// db        *sql.DB
	worlds []internal.World
}

// Initialise the server
func (server *Server) Initialise(dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.migrating = make(map[mnet.Client]bool)

	err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase)

	// common.DB, err = sql.Open("mysql", dbuser+":"+dbpassword+"@tcp("+dbaddress+":"+dbport+")/"+dbdatabase)

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

// HandleServerPacket from world
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
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

// The following logic could do with being cleaned up
func (server *Server) handleNewWorld(conn mnet.Server, reader mpacket.Reader) {
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
				if v.Conn == nil {
					server.worlds[i].Conn = conn
					name = server.worlds[i].Name

					registered = true
					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, internal.World{Conn: conn, Name: name})
			}

			p := mpacket.CreateInternal(opcode.WorldRequestOk)
			p.WriteString(name)
			conn.Send(p)

			log.Println("Registered", name)
		} else {
			registered := false
			for i, w := range server.worlds {
				if w.Name == name {
					server.worlds[i].Conn = conn
					server.worlds[i].Name = name

					p := mpacket.CreateInternal(opcode.WorldRequestOk)
					p.WriteString(name)
					conn.Send(p)

					registered = true

					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, internal.World{Conn: conn, Name: name})

				p := mpacket.CreateInternal(opcode.WorldRequestOk)
				p.WriteString(server.worlds[len(server.worlds)-1].Name)
				conn.Send(p)
			}

			log.Println("Re-registered", name)
		}
	}
}

func (server *Server) handleWorldInfo(conn mnet.Server, reader mpacket.Reader) {
	for i, v := range server.worlds {
		if v.Conn != conn {
			continue
		}

		server.worlds[i].SerialisePacket(reader)

		if v.Name == "" {
			log.Println("Registerd new world", server.worlds[i].Name)
		} else {
			log.Println("Updated world info for", v.Name)
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
