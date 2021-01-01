package main

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/nx"

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
	gameState channel.Server
}

func newChannelServer(configFile string) *channelServer {
	config, dbConfig := channelConfigFromFile(configFile)

	return &channelServer{
		eRecv:    make(chan *mnet.Event),
		wRecv:    make(chan func()),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}
}

func (cs *channelServer) run() {
	log.Println("Channel Server")

	cs.establishWorldConnection()

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)
	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	start = time.Now()
	channel.PopulateDropTable("drops.json")
	elapsed = time.Since(start)
	log.Println("Loaded and parsed drop data in", elapsed)

	cs.gameState.Initialise(cs.wRecv, cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Wait()
}

func (cs *channelServer) establishWorldConnection() {

	ticker := time.NewTicker(5 * time.Second)
	for !cs.connectToWorld() {
		cs.gameState.SendCountdownToPlayers(5)
		<-ticker.C
	}
	ticker.Stop()

	ip := net.ParseIP(cs.config.ClientConnectionAddress)
	port, err := strconv.Atoi(cs.config.ListenPort)

	if err != nil {
		panic(err)
	}

	cs.gameState.RegisterWithWorld(cs.worldConn, ip.To4(), int16(port), cs.config.MaxPop)
	cs.gameState.SendCountdownToPlayers(0)
}

func (cs *channelServer) connectToWorld() bool {
	conn, err := net.Dial("tcp", cs.config.WorldAddress+":"+cs.config.WorldPort)

	if err != nil {
		log.Println("Could not connect to world server at", cs.config.WorldAddress+":"+cs.config.WorldPort)
		cs.gameState.SendLostWorldConnectionMessage()
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

		client := mnet.NewClient(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv, cs.config.Latency, cs.config.Jitter)

		go client.Reader()
		go client.Writer()

		conn.Write(packetClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
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

			switch conn := e.Conn.(type) {
			case mnet.Client:
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", conn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", conn, "disconnected")
					cs.gameState.ClientDisconnected(conn)
				case mnet.MEClientPacket:
					cs.gameState.HandleClientPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			case mnet.Server:
				switch e.Type {
				case mnet.MEServerDisconnect:
					log.Println("Server at", conn, "disconnected")
					log.Println("Attempting to re-establish world server connection")
					cs.establishWorldConnection()
				case mnet.MEServerPacket:
					cs.gameState.HandleServerPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			}
		case work, ok := <-cs.wRecv:
			if ok {
				work()
			}
		}
	}
}
