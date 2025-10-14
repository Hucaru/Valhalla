package main

import (
	"context"
	"crypto/rand"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	// graceful shutdown & listeners
	ctx            context.Context
	cancel         context.CancelFunc
	clientListener net.Listener
	serverListener net.Listener
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
	ctx, cancel := context.WithCancel(context.Background())

	return &loginServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (ls *loginServer) run() {
	log.Println("Login Server")
	log.Printf("Listening on %q:%q", ls.config.ClientListenAddress, ls.config.ClientListenPort)

	start := time.Now()
	nx.LoadFile("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data (NX) in", elapsed)

	ls.gameState.Initialise(ls.dbConfig.User, ls.dbConfig.Password, ls.dbConfig.Address, ls.dbConfig.Port, ls.dbConfig.Database, ls.config.WithPin, ls.config.AutoRegister)

	// OS signal handler for graceful shutdown
	ls.wg.Add(1)
	go func() {
		defer ls.wg.Done()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			log.Println("Shutdown signal received")
			ls.shutdown()
		case <-ls.ctx.Done():
		}
	}()

	ls.wg.Add(1)
	go ls.acceptNewClientConnections()

	ls.wg.Add(1)
	go ls.acceptNewServerConnections()

	ls.wg.Add(1)
	go ls.processEvent()

	// Block until all goroutines exit
	ls.wg.Wait()
	log.Println("Login Server stopped")
}

// shutdown triggers graceful stop: stop accepting, then let workers drain
func (ls *loginServer) shutdown() {
	// Cancel context so loops can exit
	ls.cancel()

	// Close listeners to unblock Accept()
	if ls.clientListener != nil {
		_ = ls.clientListener.Close()
	}
	if ls.serverListener != nil {
		_ = ls.serverListener.Close()
	}

	// Note: we intentionally do NOT close ls.eRecv here.
	// Existing mnet readers/writers may still attempt to publish.
	// processEvent will exit via ctx.Done().
}

func isTempNetErr(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout() || ne.Temporary()
	}
	return false
}

func (ls *loginServer) acceptNewServerConnections() {
	defer ls.wg.Done()

	l, err := net.Listen("tcp", ls.config.ServerListenAddress+":"+ls.config.ServerListenPort)
	if err != nil {
		log.Println("server listen error:", err)
		// If we cannot listen at all, cancel the server
		ls.shutdown()
		return
	}
	ls.serverListener = l
	log.Println("Server listener ready:", ls.config.ServerListenAddress+":"+ls.config.ServerListenPort)
	defer func() {
		_ = l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ls.ctx.Err() != nil {
				// shutting down
				return
			}
			if isTempNetErr(err) {
				log.Println("temporary server Accept error:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("fatal server Accept error:", err)
			// Do not close eRecv; just stop this accept loop.
			return
		}

		serverConn := mnet.NewServer(conn, ls.eRecv, ls.config.PacketQueueSize)
		go serverConn.Reader()
		go serverConn.Writer()
	}
}

func (ls *loginServer) acceptNewClientConnections() {
	defer ls.wg.Done()

	l, err := net.Listen("tcp", ls.config.ClientListenAddress+":"+ls.config.ClientListenPort)
	if err != nil {
		log.Println("client listen error:", err)
		ls.shutdown()
		return
	}
	ls.clientListener = l
	log.Println("Client listener ready:", ls.config.ClientListenAddress+":"+ls.config.ClientListenPort)
	defer func() {
		_ = l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ls.ctx.Err() != nil {
				// shutting down
				return
			}
			if isTempNetErr(err) {
				log.Println("temporary client Accept error:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("fatal client Accept error:", err)
			// Do not close eRecv; just stop this accept loop.
			return
		}

		// Handshake keys
		keySend := [4]byte{}
		_, _ = rand.Read(keySend[:])
		keyRecv := [4]byte{}
		_, _ = rand.Read(keyRecv[:])

		client := mnet.NewClient(conn, ls.eRecv, ls.config.PacketQueueSize, keySend, keyRecv, ls.config.Latency, ls.config.Jitter)
		go client.Reader()
		go client.Writer()

		// Initial handshake
		_, _ = conn.Write(packetClientHandshake(constant.MapleVersion, keyRecv[:], keySend[:]))
	}
}

func (ls *loginServer) processEvent() {
	defer ls.wg.Done()

	for {
		select {
		case <-ls.ctx.Done():
			log.Println("Stopping event handling: shutdown")
			return
		case e, ok := <-ls.eRecv:
			if !ok {
				// We never close eRecv during normal runtime; this would indicate upstream closed it.
				log.Println("Stopping event handling: event channel closed")
				return
			}
			switch conn := e.Conn.(type) {
			case mnet.Client:
				switch e.Type {
				case mnet.MEClientConnected:
					log.Println("New client from", conn)
				case mnet.MEClientDisconnect:
					log.Println("Client at", conn, "disconnected")
					// Ensure game state cleanup on disconnect
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
			default:
				// Unknown event origin; ignore safely
			}
		}
	}
}
