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

	// graceful shutdown & listener
	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
}

func newChannelServer(configFile string) *channelServer {
	config, dbConfig := channelConfigFromFile(configFile)
	ctx, cancel := context.WithCancel(context.Background())

	return &channelServer{
		eRecv:    make(chan *mnet.Event),
		wRecv:    make(chan func()),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (cs *channelServer) run() {
	log.Println("Channel Server")

	// Signal handler for graceful shutdown
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

	cs.establishWorldConnection()

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)
	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	start = time.Now()
	channel.PopulateDropTable("drops.json")
	elapsed = time.Since(start)
	log.Println("Loaded and parsed drop data in", elapsed)

	channel.StartSaver()
	defer channel.StopSaver()

	cs.gameState.Initialise(cs.wRecv, cs.dbConfig.User, cs.dbConfig.Password, cs.dbConfig.Address, cs.dbConfig.Port, cs.dbConfig.Database)

	cs.wg.Add(1)
	go cs.acceptNewConnections()

	cs.wg.Add(1)
	go cs.processEvent()

	cs.wg.Wait()
	log.Println("Channel Server stopped")
}

func (cs *channelServer) shutdown() {
	// Cancel context so loops can exit
	cs.cancel()
	// Close listener to unblock Accept()
	if cs.listener != nil {
		_ = cs.listener.Close()
	}
	// Note: we intentionally do NOT close cs.eRecv or cs.wRecv here since
	// other goroutines may still attempt to publish work/events.
}

func (cs *channelServer) establishWorldConnection() {
	// Context-aware exponential backoff retry
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
			cs.gameState.RegisterWithWorld(cs.worldConn, ip.To4(), int16(port), cs.config.MaxPop)
			cs.gameState.SendCountdownToPlayers(0)
			return
		}
		// If connection failed, notify players and back off
		cs.gameState.SendLostWorldConnectionMessage()
		cs.gameState.SendCountdownToPlayers(int32(backoff / time.Second))
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
			if backoff > 10*time.Second {
				backoff = 10 * time.Second
			}
		}
	}
}

func (cs *channelServer) connectToWorld() bool {
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

func (cs *channelServer) acceptNewConnections() {
	defer cs.wg.Done()

	l, err := net.Listen("tcp", cs.config.ListenAddress+":"+cs.config.ListenPort)
	if err != nil {
		log.Println("channel listen error:", err)
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
				// shutting down
				return
			}
			if isTempNetErr(err) {
				log.Println("temporary client Accept error:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("fatal client Accept error:", err)
			// Do not close cs.eRecv; just stop this accept loop
			return
		}

		keySend := [4]byte{}
		_, _ = rand.Read(keySend[:])
		keyRecv := [4]byte{}
		_, _ = rand.Read(keyRecv[:])

		client := mnet.NewClient(conn, cs.eRecv, cs.config.PacketQueueSize, keySend, keyRecv, cs.config.Latency, cs.config.Jitter)

		go client.Reader()
		go client.Writer()

		// Best-effort initial handshake
		if _, err := conn.Write(packetClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:])); err != nil {
			// If this fails immediately, let the client's goroutines handle cleanup via disconnect
			log.Println("handshake write failed:", err)
		}
	}
}

func (cs *channelServer) processEvent() {
	defer cs.wg.Done()

	for {
		select {
		case <-cs.ctx.Done():
			log.Println("Stopping event handling: shutdown")
			return

		case e, ok := <-cs.eRecv:
			if !ok {
				log.Println("Stopping event handling due to event channel close")
				return
			}

			switch conn := e.Conn.(type) {
			case mnet.Client:
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", conn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", conn, "disconnected")

					// Ensure we persist the player on socket close.
					channel.OnDisconnect(conn)
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
