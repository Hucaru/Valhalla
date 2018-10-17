package server

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"

	"github.com/Hucaru/Valhalla/handlers/loginhandlers"
	"github.com/Hucaru/Valhalla/maplepacket"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
)

type loginServer struct {
	config   loginConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup
}

func NewLoginServer(configFile string) *loginServer {
	config, dbConfig := loginConfigFromFile("config.toml")

	ls := &loginServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	return ls
}

func (ls *loginServer) Run() {
	log.Println("LoginServer")

	ls.wg.Add(1)
	go ls.establishDatabaseConnection()

	ls.wg.Add(1)
	go ls.acceptNewConnections()

	ls.wg.Add(1)
	go ls.processEvent()

	ls.wg.Wait()
}

func (ls *loginServer) establishDatabaseConnection() {
	database.Connect(ls.dbConfig.User, ls.dbConfig.Password, ls.dbConfig.Address, ls.dbConfig.Port, ls.dbConfig.Database)
}

func (ls *loginServer) acceptNewConnections() {
	defer ls.wg.Done()

	listener, err := net.Listen("tcp", ls.config.ListenAddress+":"+ls.config.ListenPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Client listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
			close(ls.eRecv)
			return
		}

		keySend := [4]byte{}
		rand.Read(keySend[:])
		keyRecv := [4]byte{}
		rand.Read(keyRecv[:])

		loginConn := mnet.NewLogin(conn, ls.eRecv, ls.config.PacketQueueSize, keySend, keyRecv)

		go loginConn.Reader()
		go loginConn.Writer()

		conn.Write(packets.ClientHandshake(consts.MapleVersion, keyRecv[:], keySend[:]))
	}
}

func (ls *loginServer) processEvent() {
	defer ls.wg.Done()

	for {
		select {
		case e, ok := <-ls.eRecv:

			if !ok {
				log.Println("Stopping event handling due to channel error")
				return
			}

			loginConn, ok := e.Conn.(mnet.MConnLogin)

			if !ok {
				log.Fatal("Error in converting MConn to MConnLogin")
			}

			switch e.Type {
			case mnet.MEClientConnected:
				log.Println("New client from", loginConn)
			case mnet.MEClientDisconnect:
				log.Println("Client at", loginConn, "disconnected")
				loginConn.Cleanup()
			case mnet.MEClientPacket:
				loginhandlers.HandlePacket(loginConn, maplepacket.NewReader(&e.Packet))
			}
		}
	}
}
