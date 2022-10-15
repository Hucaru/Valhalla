package main

import (
	"github.com/Hucaru/Valhalla/internal"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/world"
)

type worldServer struct {
	config   worldConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup
	lconn    mnet.Server
	state    world.Server
}

func newWorldServer(configFile string) *worldServer {
	config, dbConfig := worldConfigFromFile(configFile)
	ws := worldServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	ws.state.Info.DefaultRates = internal.Rates{Exp: config.ExpRate, Drop: config.DropRate, Mesos: config.MesosRate}
	ws.state.Info.Rates = ws.state.Info.DefaultRates
	ws.state.Info.Ribbon = config.Ribbon
	ws.state.Info.Message = config.Message

	return &ws
}

func (ws *worldServer) run() {
	log.Println("World Server")

	ws.establishLoginConnection()

	ws.wg.Add(1)
	go ws.acceptNewServerConnections()

	ws.wg.Add(1)
	go ws.processEvent()

	ws.wg.Wait()
}

func (ws *worldServer) establishLoginConnection() {
	ticker := time.NewTicker(5 * time.Second)
	for !ws.connectToLogin() {
		<-ticker.C
	}
	ticker.Stop()

	ws.state.RegisterWithLogin(ws.lconn)
}

func (ws *worldServer) connectToLogin() bool {
	conn, err := net.Dial("tcp", ws.config.LoginAddress+":"+ws.config.LoginPort)

	if err != nil {
		log.Println("Could not connect to login server at", ws.config.LoginAddress+":"+ws.config.LoginPort)
		return false
	}

	log.Println("Connected to login server at", ws.config.LoginAddress+":"+ws.config.LoginPort)

	login := mnet.NewServer(conn, ws.eRecv, ws.config.PacketQueueSize)

	go login.Reader()
	go login.Writer()

	ws.lconn = login

	return true
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

			switch conn := e.Conn.(type) {
			case mnet.Server:
				switch e.Type {
				case mnet.MEServerConnected:
					log.Println("New server from", conn)
				case mnet.MEServerDisconnect:
					log.Println("Server at", conn, "disconnected")

					if conn == ws.lconn {
						log.Println("Attempting to re-establish login server connection")
						ws.establishLoginConnection()
					}

					ws.state.ServerDisconnected(conn)
				case mnet.MEServerPacket:
					ws.state.HandleServerPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			}

		}

	}
}
