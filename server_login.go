package main

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/login"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/mnet"
)

type loginServer struct {
	config    loginConfig
	dbConfig  dbConfig
	eRecv     chan *mnet.Event
	wg        *sync.WaitGroup
	gameState login.Server
}

func packetClientHandshake(mapleVersion int16, recv, send []byte) mpacket.Packet {
	p := mpacket.NewPacket()

	p.WriteInt16(13)
	p.WriteInt16(mapleVersion)
	p.WriteString("")
	p.Append(recv)
	p.Append(send)
	p.WriteByte(8)

	return p

}

func newLoginServer(configFile string) *loginServer {
	config, dbConfig := loginConfigFromFile(configFile)

	return &loginServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}
}

func (ls *loginServer) run() {
	log.Println("Login Server")

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	ls.gameState.Initialise(ls.dbConfig.User, ls.dbConfig.Password, ls.dbConfig.Address, ls.dbConfig.Port, ls.dbConfig.Database, ls.config.WithPin)

	ls.wg.Add(1)
	go ls.acceptNewClientConnections()

	ls.wg.Add(1)
	go ls.acceptNewServerConnections()

	ls.wg.Add(1)
	go ls.processEvent()

	ls.wg.Wait()
}

func (ls *loginServer) acceptNewServerConnections() {
	defer ls.wg.Done()

	listener, err := net.Listen("tcp", ls.config.ServerListenAddress+":"+ls.config.ServerListenPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Server listener ready:", ls.config.ServerListenAddress+":"+ls.config.ServerListenPort)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
			close(ls.eRecv)
			return
		}

		serverConn := mnet.NewServer(conn, ls.eRecv, ls.config.PacketQueueSize)

		go serverConn.Reader()
		go serverConn.Writer()
	}
}

func (ls *loginServer) acceptNewClientConnections() {
	defer ls.wg.Done()

	listener, err := net.Listen("tcp", ls.config.ClientListenAddress+":"+ls.config.ClientListenPort)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("Client listener ready:", ls.config.ClientListenAddress+":"+ls.config.ClientListenPort)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
			close(ls.eRecv)
			return
		}

		// ls.gameState.ClientConnected(conn, ls.eRecv, ls.config.PacketQueueSize)
		keySend := [4]byte{}
		rand.Read(keySend[:])
		keyRecv := [4]byte{}
		rand.Read(keyRecv[:])

		client := mnet.NewClient(conn, ls.eRecv, ls.config.PacketQueueSize, keySend, keyRecv, ls.config.Latency, ls.config.Jitter)

		go client.Reader()
		go client.Writer()

		conn.Write(packetClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
	}
}

func (ls *loginServer) processEvent() {
	defer ls.wg.Done()

	for {
		select {
		case e, ok := <-ls.eRecv:

			if !ok {
				log.Println("Stopping event handling due to channel read error")
				return
			}

			switch conn := e.Conn.(type) {
			case mnet.Client:
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", conn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", conn, "disconnected")
					ls.gameState.ClientDisconnected(conn)
				case mnet.MEClientPacket:
					ls.gameState.HandleClientPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			case mnet.Server:
				switch e.Type {
				case mnet.MEServerConnected:
					log.Println("New server from", conn)
				case mnet.MEServerDisconnect:
					log.Println("Server at", conn, "disconnected")
					ls.gameState.ServerDisconnected(conn)
				case mnet.MEServerPacket:
					ls.gameState.HandleServerPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			}
		}
	}
}
