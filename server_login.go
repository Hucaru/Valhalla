package main

import (
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server"

	"github.com/Hucaru/Valhalla/mnet"
)

type loginServer struct {
	config    loginConfig
	dbConfig  dbConfig
	eRecv     chan *mnet.Event
	wg        *sync.WaitGroup
	gameState server.LoginServer
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

	ls.gameState.Initialise(ls.dbConfig.User, ls.dbConfig.Password, ls.dbConfig.Address, ls.dbConfig.Port, ls.dbConfig.Database)

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

		client := mnet.NewClient(conn, ls.eRecv, ls.config.PacketQueueSize, keySend, keyRecv)

		go client.Reader()
		go client.Writer()

		conn.Write(server.PacketClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
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

			clientConn, ok := e.Conn.(mnet.Client)

			if ok {
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", clientConn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", clientConn, "disconnected")
					ls.gameState.ClientDisconnected(clientConn)
				case mnet.MEClientPacket:
					ls.gameState.HandleClientPacket(clientConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			} else {
				serverConn, ok := e.Conn.(mnet.Server)

				if ok {
					switch e.Type {
					case mnet.MEServerConnected:
						log.Println("New server from", serverConn)
					case mnet.MEServerDisconnect:
						log.Println("Server at", serverConn, "disconnected")
						ls.gameState.ServerDisconnected(serverConn)
					case mnet.MEServerPacket:
						ls.gameState.HandleServerPacket(serverConn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
					}
				}
			}

		}
	}
}