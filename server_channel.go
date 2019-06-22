package main

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/server"
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
	gameState server.ChannelServer
}

func newChannelServer(configFile string) *channelServer {
	config, dbConfig := channelConfigFromFile(configFile)

	cs := &channelServer{
		eRecv:    make(chan *mnet.Event),
		wRecv:    make(chan func()),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	return cs
}

func (cs *channelServer) run() {
	log.Println("Channel Server")

	cs.establishWorldConnection()

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

func (cs *channelServer) establishWorldConnection() {
	ticker := time.NewTicker(5 * time.Second)
	for !cs.connectToWorld() {
		<-ticker.C
	}
	ticker.Stop()

	ip := net.ParseIP(cs.config.ClientConnectionAddress)
	port, err := strconv.Atoi(cs.config.ListenPort)

	if err != nil {
		panic(err)
	}

	cs.gameState.RegisterWithWorld(cs.worldConn, ip.To4(), int16(port), cs.config.MaxPop)
}

func (cs *channelServer) connectToWorld() bool {
	conn, err := net.Dial("tcp", cs.config.WorldAddress+":"+cs.config.WorldPort)

	if err != nil {
		log.Println("Could not connect to world server at", cs.config.WorldAddress+":"+cs.config.WorldPort)
		return false
	}

	log.Println("Connected to world server at", cs.config.WorldAddress+":"+cs.config.WorldPort)

	world := mnet.NewServer(conn, cs.eRecv, cs.config.PacketQueueSize)

	go world.Reader()
	go world.Writer()

	cs.worldConn = world

	return true
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

		keySend := [4]byte{}
		rand.Read(keySend[:])
		keyRecv := [4]byte{}
		rand.Read(keyRecv[:])

		client := mnet.NewClient(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv)

		go client.Reader()
		go client.Writer()

		conn.Write(server.PacketClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
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
						log.Println("Attempting to re-establish world server connection")
						cs.establishWorldConnection()
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
