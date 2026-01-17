package main

import (
	"context"
	"crypto/rand"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Hucaru/Valhalla/cashshop"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type cashShopServer struct {
	config    cashShopConfig
	dbConfig  dbConfig
	eRecv     chan *mnet.Event
	wRecv     chan func()
	wg        *sync.WaitGroup
	worldConn mnet.Server
	gameState cashshop.Server

	ctx           context.Context
	cancel        context.CancelFunc
	listener      net.Listener
	dispatchReady chan struct{}
}

func newCashShopServer(configFile string) *cashShopServer {
	config, dbConfig := cashShopConfigFromFile(configFile)
	ctx, cancel := context.WithCancel(context.Background())

	return &cashShopServer{
		eRecv:         make(chan *mnet.Event),
		wRecv:         make(chan func()),
		config:        config,
		dbConfig:      dbConfig,
		wg:            &sync.WaitGroup{},
		ctx:           ctx,
		cancel:        cancel,
		dispatchReady: make(chan struct{}),
	}
}

func (cs *cashShopServer) run() {
	log.Println("CashShop Server")
	log.Printf("Listening on %q:%q", cs.config.ListenAddress, cs.config.ListenPort)

	cs.wg.Add(1)
	go func() {
		defer cs.wg.Done()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			log.Println("Shutdown signal received")
			cs.shutdown()
		case <-cs.ctx.Done():
		}
	}()

	cs.wg.Add(1)
	go cs.processEvent()

	<-cs.dispatchReady

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)
	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	cs.gameState.Initialise(cs.wRecv, cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.establishWorldConnection()
	cs.gameState.StartAutosave(cs.ctx)

	cs.wg.Wait()
	log.Println("CashShop Server stopped")
}

func (cs *cashShopServer) shutdown() {
	cs.gameState.CheckpointAll(cs.ctx)

	cs.cancel()
	if cs.listener != nil {
		_ = cs.listener.Close()
	}
}

func (cs *cashShopServer) establishWorldConnection() {
	backoff := time.Second
	for {
		select {
		case <-cs.ctx.Done():
			return
		default:
		}
		if cs.connectToWorld() {
			ip := net.ParseIP(cs.config.ClientConnectionAddress)
			port, err := strconv.Atoi(cs.config.ListenPort)
			if err != nil {
				log.Println("invalid listen port:", err)
				return
			}

			cs.wRecv <- func() {
				cs.gameState.RegisterWithWorld(cs.worldConn, ip.To4(), int16(port))
			}
			return
		}

		cs.wRecv <- func() {
			// Lost Connection
			// Handle?
		}
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
			if backoff > 10*time.Second {
				backoff = 10 * time.Second
			}
		}
	}
}

func (cs *cashShopServer) connectToWorld() bool {
	addr := cs.config.WorldAddress + ":" + cs.config.WorldPort
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Could not connect to world server at", addr)
		return false
	}

	log.Println("Connected to world server at", addr)

	world := mnet.NewServer(conn, cs.eRecv, cs.config.PacketQueueSize)

	go world.Reader()
	go world.Writer()

	cs.worldConn = world
	return true
}

func (cs *cashShopServer) acceptNewConnections() {
	defer cs.wg.Done()

	l, err := net.Listen("tcp", cs.config.ListenAddress+":"+cs.config.ListenPort)
	if err != nil {
		log.Println("cashshop listen error:", err)
		cs.shutdown()
		return
	}
	cs.listener = l
	log.Println("Client listener ready:", cs.config.ListenAddress+":"+cs.config.ListenPort)
	defer func() { _ = l.Close() }()

	for {
		conn, err := l.Accept()
		if err != nil {
			if cs.ctx.Err() != nil {
				return
			}
			if isTempNetErr(err) {
				log.Println("temporary client Accept error:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("fatal client Accept error:", err)
			return
		}

		keySend := [4]byte{}
		_, _ = rand.Read(keySend[:])
		keyRecv := [4]byte{}
		_, _ = rand.Read(keyRecv[:])

		client := mnet.NewClient(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv, cs.config.Latency, cs.config.Jitter)

		go client.Reader()
		go client.Writer()

		if _, err := conn.Write(packetClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:])); err != nil {
			log.Println("handshake write failed:", err)
		}

	}
}

func (cs *cashShopServer) processEvent() {
	defer cs.wg.Done()

	select {
	case <-cs.ctx.Done():
		return
	default:
	}
	if cs.dispatchReady != nil {
		close(cs.dispatchReady)
		cs.dispatchReady = nil
	}

	for {
		select {
		case <-cs.ctx.Done():
			log.Println("Stopping event handling: shutdown")
			return

		case e, ok := <-cs.eRecv:
			if !ok {
				log.Println("Stopping event handling due to event cashshop close")
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
					go cs.establishWorldConnection()

				case mnet.MEServerPacket:
					cs.gameState.HandleServerPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			}

		case work, ok := <-cs.wRecv:
			if !ok {
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("panic in scheduled work:", r)
					}
				}()
				work()
			}()
		}
	}
}
