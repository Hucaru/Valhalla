package server

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/game/script"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type channelServer struct {
	config    channelConfig
	dbConfig  dbConfig
	eRecv     chan *mnet.Event
	wRecv     chan func()
	wg        *sync.WaitGroup
	worldConn mnet.Server
	gameState game.Channel
}

func NewChannelServer(configFile string) *channelServer {
	config, dbConfig := channelConfigFromFile("config.toml")

	cs := &channelServer{
		eRecv:    make(chan *mnet.Event),
		wRecv:    make(chan func()),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	return cs
}

func (cs *channelServer) Run() {
	log.Println("Channel Server")

	cs.connectToWorld()

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	cs.gameState.Initialise(cs.wRecv, cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)

	go script.WatchScriptDirectory("scripts/npc/")
	go script.WatchScriptDirectory("scripts/event/")
	go script.WatchScriptDirectory("scripts/admin/")

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Wait()
}

func (cs *channelServer) connectToWorld() {
	conn, err := net.Dial("tcp", cs.config.WorldAddress+":"+cs.config.WorldPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Connected to world server at", cs.config.WorldAddress+":"+cs.config.WorldPort)

	ip := net.ParseIP(cs.config.ClientConnectionAddress)
	port, err := strconv.Atoi(cs.config.ListenPort)

	if err != nil {
		panic(err)
	}

	world := mnet.NewServer(conn, cs.eRecv, cs.config.PacketQueueSize)

	go world.Reader()
	go world.Writer()

	cs.worldConn = world
	cs.gameState.RegisterWithWorld(cs.worldConn, ip.To4(), int16(port))
}

func (cs *channelServer) acceptNewConnections() {
	defer cs.wg.Done()

	listener, err := net.Listen("tcp", cs.config.ListenAddress+":"+cs.config.ListenPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Client listener ready:", cs.config.ListenAddress+":"+cs.config.ListenPort)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
			close(cs.eRecv)
			return
		}

		cs.gameState.ClientConnected(conn, cs.eRecv, cs.config.PacketQueueSize)
	}
}

func (cs *channelServer) processEvent() {
	defer cs.wg.Done()

	for {
		select {
		case e, ok := <-cs.eRecv:

			if !ok {
				log.Println("Stopping event handling due to channel error")
				return
			}

			clientConn, ok := e.Conn.(mnet.Client)

			if ok {
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", clientConn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", clientConn, "disconnected")
					cs.gameState.ClientDisconnected(clientConn)
				case mnet.MEClientPacket:
					cs.gameState.HandleClientPacket(clientConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			} else {
				serverConn, ok := e.Conn.(mnet.Server)

				if ok {
					switch e.Type {
					case mnet.MEServerDisconnect:
						log.Println("Server at", serverConn, "disconnected")
					case mnet.MEServerPacket:
						cs.gameState.HandleServerPacket(serverConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
					}
				}
			}
		case work, ok := <-cs.wRecv:
			if ok {
				work()
			}
		}
	}
}
