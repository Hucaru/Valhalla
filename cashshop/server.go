package cashshop

import (
	"context"
	"log"
	"time"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Server state
type Server struct {
	id        byte
	worldName string
	dispatch  chan func()
	world     mnet.Server
	ip        []byte
	port      int16
	migrating []mnet.Client
	players   channel.Players
	channels  [20]internal.Channel
}

// Initialise the server
func (server *Server) Initialise(work chan func(), dbuser, dbpassword, dbaddress, dbport, dbdatabase string) {
	server.dispatch = work
	server.id = 50
	server.players = channel.NewPlayers()

	if err := common.ConnectToDB(dbuser, dbpassword, dbaddress, dbport, dbdatabase); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")

	log.Println("Initialised game state")

	common.StartMetrics()
	log.Println("Started serving metrics on :" + common.MetricsPort)
}

// RegisterWithWorld server
func (server *Server) RegisterWithWorld(conn mnet.Server, ip []byte, port int16) {
	server.world = conn
	server.ip = ip
	server.port = port

	server.registerWithWorld()
}

func (server *Server) registerWithWorld() {
	p := mpacket.CreateInternal(opcode.CashShopNew)
	p.WriteBytes(server.ip)
	p.WriteInt16(server.port)
	server.world.Send(p)
}

// ClientDisconnected from server
func (server *Server) ClientDisconnected(conn mnet.Client) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	plr.Logout()

	if remPlrErr := server.players.RemoveFromConn(conn); remPlrErr != nil {
		log.Println(remPlrErr)
	}

	if idx := func() int {
		for i, v := range server.migrating {
			if v == conn {
				return i
			}
		}
		return -1
	}(); idx > -1 {
		server.migrating = append(server.migrating[:idx], server.migrating[idx+1:]...)
	}

	if _, dbErr := common.DB.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID()); dbErr != nil {
		log.Println("Unable to complete logout for ", conn.GetAccountID())
	}

	server.world.Send(internal.PacketChannelPlayerDisconnect(plr.ID, plr.Name, 0))
}

// CheckpointAll now uses the saver to flush debounced/coalesced deltas for every player.
func (server *Server) CheckpointAll() {
	if server.dispatch == nil {
		return
	}
	done := make(chan struct{})
	server.dispatch <- func() {
		server.players.Flush()
		close(done)
	}
	<-done
}

// startAutosave periodically flushes deltas via the saver.
func (server *Server) StartAutosave(ctx context.Context) {
	if server.dispatch == nil {
		return
	}
	const interval = 30 * time.Second

	var scheduleNext func()
	scheduleNext = func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.AfterFunc(interval, func() {
			server.dispatch <- func() {
				server.players.Flush()
				scheduleNext()
			}
		})
	}

	server.dispatch <- func() {
		server.players.Flush()
		scheduleNext()
	}
}
