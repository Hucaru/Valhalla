package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/Hucaru/Valhalla/constant"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Hucaru/Valhalla/channel"
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
	config, dbConfig := loadChannelConfig(configFile)
	return &channelServer{
		eRecv:    make(chan *mnet.Event, 4096*4),
		wRecv:    make(chan func(), 4096*4),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
	}
}

func (cs *channelServer) run() {
	log.Println("Channel Server")

	//cs.establishWorldConnection()

	start := time.Now()
	//nx.LoadFile("Data.nx")
	elapsed := time.Since(start)
	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	start = time.Now()
	channel.PopulateDropTable("drops.json")
	elapsed = time.Since(start)
	log.Println("Loaded and parsed drop data in", elapsed)

	cs.gameState.Initialize(cs.wRecv, cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Add(1)
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGKILL)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		fmt.Println("Wait for 5 second to finish processing")
		cs.wg.Done()
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

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

		client := mnet.NewClientMeta(conn, cs.config.PacketQueueSize, cs.config.Latency, cs.config.Jitter)
		cs.gameState.HandleClientPacket(client, mpacket.Reader{}, constant.OnConnected)

		go func() {
			for {
				buff := make(mpacket.Packet, constant.MetaClientHeaderSize)
				if _, err := conn.Read(buff); err == io.EOF || err != nil {
					fmt.Println("Error reading:", err.Error())
					cs.gameState.HandleClientPacket(client, mpacket.Reader{}, constant.OnDisconnected)
					return
				}

				msgLen := binary.BigEndian.Uint32(buff[:4])
				msgProtocol := binary.BigEndian.Uint32(buff[4:8])

				buff = make([]byte, msgLen)
				if _, err := conn.Read(buff); err == io.EOF || err != nil {
					fmt.Println("Error reading:", err.Error())
					cs.gameState.HandleClientPacket(client, mpacket.Reader{}, constant.OnDisconnected)
					return
				}
				cs.gameState.HandleClientPacket(client, mpacket.NewReader(&buff, time.Now().Unix()), msgProtocol)
			}
		}()

		//go client.MetaWriter()
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
			fmt.Println("WORK", work)
			if ok {
				work()
			}
		}

	}
}
