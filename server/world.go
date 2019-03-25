package server

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type worldServer struct {
	config   worldConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup
	lconn    mnet.Server
	state    game.World
}

func NewWorldServer(configFile string) *worldServer {
	config, dbConfig := worldConfigFromFile("config.toml")

	ws := &worldServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	return ws
}

func (ws *worldServer) Run() {
	log.Println("World Server")

	ws.connectToLogin()

	ws.wg.Add(1)
	go ws.acceptNewServerConnections()

	ws.wg.Add(1)
	go ws.processEvent()

	ws.wg.Wait()
}

func (ws *worldServer) connectToLogin() {
	conn, err := net.Dial("tcp", ws.config.LoginAddress+":"+ws.config.LoginPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Connected to login server at", ws.config.LoginAddress+":"+ws.config.LoginPort)

	ws.lconn = mnet.NewServer(conn, ws.eRecv, ws.config.PacketQueueSize)
}

func (ws *worldServer) acceptNewServerConnections() {
	defer ws.wg.Done()

	listener, err := net.Listen("tcp", ws.config.ListenAddress+":"+ws.config.ListenPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Server listener ready:", ws.config.ListenAddress+":"+ws.config.ListenPort)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
			close(ws.eRecv)
			return
		}

		serverConn := mnet.NewServer(conn, ws.eRecv, ws.config.PacketQueueSize)

		go serverConn.Reader()
		go serverConn.Writer()
	}
}

func (ws *worldServer) processEvent() {
	defer ws.wg.Done()

	for {
		select {
		case e, ok := <-ws.eRecv:

			if !ok {
				log.Println("Stopping event handling due to channel read error")
				return
			}

			serverConn, ok := e.Conn.(mnet.Server)

			if !ok {
				panic("Invalid type assestion")
			}

			switch e.Type {
			case mnet.MEServerConnected:
				log.Println("New server from", serverConn)
			case mnet.MEServerDisconnect:
				log.Println("Server at", serverConn, "disconnected")
			case mnet.MEServerPacket:
				ws.state.HandleChannelPacket(serverConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
			}
		}

	}
}
