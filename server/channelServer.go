package server

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/handlers/channelhandlers"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
)

type channelServer struct {
	config   channelConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup
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

	cs.wg.Add(1)
	go cs.establishDatabaseConnection()

	start := time.Now()
	nx.Parse("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	game.InitMaps()

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Wait()
}

func (cs *channelServer) establishDatabaseConnection() {
	database.Connect(cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)
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

		loginConn := mnet.NewLogin(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv)

		go loginConn.Reader()
		go loginConn.Writer()

		conn.Write(packets.ClientHandshake(consts.MapleVersion, keyRecv[:], keySend[:]))
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

			if !ok {
				log.Fatal("Error in converting MConn to MConnChannel")
			}

			switch e.Type {
			case mnet.MEClientConnected:
				log.Println("New client from", channelConn)
			case mnet.MEClientDisconnect:
				log.Println("Client at", channelConn, "disconnected")
				channelConn.Cleanup()
				game.RemovePlayer(channelConn)
			case mnet.MEClientPacket:
				channelhandlers.HandlePacket(channelConn, maplepacket.NewReader(&e.Packet))
			}
		}
	}
}
