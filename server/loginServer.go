package server

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/mnet"
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

		loginConn := mnet.NewLogin(conn, ls.eRecv, ls.config.PacketQueueSize)

		go loginConn.Reader() // reads packets from client and generates and dispatches events to the processEvent thread
		go loginConn.Writer() // listens on a buffered channel to write packets to client.
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
				// go off to handler
			}
		}
	}
}
