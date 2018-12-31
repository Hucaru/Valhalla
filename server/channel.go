package server

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/npcchat"
	"github.com/Hucaru/Valhalla/handlers/channel"
	"github.com/Hucaru/Valhalla/handlers/world"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/game/script"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type channelServer struct {
	config   channelConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup
	wconn    net.Conn
}

func NewChannelServer(configFile string) *channelServer {
	config, dbConfig := channelConfigFromFile("config.toml")

	cs := &channelServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}

	return cs
}

func (cs *channelServer) Run() {
	log.Println("Channel Server")

	cs.establishDatabaseConnection()
	cs.connectToWorld()

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	game.InitMaps()

	go script.WatchScriptDirectory("scripts/npc/")
	go script.WatchScriptDirectory("scripts/event/")

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Wait()
}

func (cs *channelServer) establishDatabaseConnection() {
	database.Connect(cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)
	go database.Monitor()
}

func (cs *channelServer) connectToWorld() {
	conn, err := net.Dial("tcp", cs.config.WorldAddress+":"+cs.config.WorldPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Connected to world server at", cs.config.WorldAddress+":"+cs.config.WorldPort)

	cs.wconn = conn
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

		channelConn := mnet.NewChannel(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv)

		go channelConn.Reader()
		go channelConn.Writer()

		conn.Write(packet.ClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
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

			channelConn, ok := e.Conn.(mnet.MConnChannel)

			if ok {
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", channelConn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", channelConn, "disconnected")
					npcchat.RemoveSession(channelConn)
					game.RemovePlayer(channelConn)
					channelConn.Cleanup()
				case mnet.MEClientPacket:
					channel.HandlePacket(channelConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			} else {
				serverConn, ok := e.Conn.(mnet.MConnServer)

				if ok {
					switch e.Type {
					case mnet.MEServerDisconnect:
						log.Println("Server at", serverConn, "disconnected")
					case mnet.MEServerPacket:
						world.HandlePacket(nil, mpacket.NewReader(&e.Packet, time.Now().Unix()))
					}
				}
			}

		}
	}
}
