package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mpacket"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/world"
)

type worldServer struct {
	config   worldConfig
	dbConfig dbConfig
	eRecv    chan *mnet.Event
	wg       *sync.WaitGroup

	lconn mnet.Server
	state world.Server

	// graceful shutdown & listener
	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
}

func newWorldServer(configFile string) *worldServer {
	config, dbConfig := worldConfigFromFile(configFile)
	ctx, cancel := context.WithCancel(context.Background())

	ws := worldServer{
		eRecv:    make(chan *mnet.Event),
		config:   config,
		dbConfig: dbConfig,
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		cancel:   cancel,
	}

	ws.state.Info.DefaultRates = internal.Rates{Exp: config.ExpRate, Drop: config.DropRate, Mesos: config.MesosRate}
	ws.state.Info.Rates = ws.state.Info.DefaultRates
	ws.state.Info.Ribbon = config.Ribbon
	ws.state.Info.Message = config.Message

	return &ws
}

func (ws *worldServer) run() {
	log.Println("World Server")
	log.Printf("Listening on %q:%q", ws.config.ListenAddress, ws.config.ListenPort)

	ws.state.Initialise(ws.dbConfig.User, ws.dbConfig.Password, ws.dbConfig.Address, ws.dbConfig.Port, ws.dbConfig.Database)

	// Signal handler for graceful shutdown
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			log.Println("Shutdown signal received")
			ws.shutdown()
		case <-ws.ctx.Done():
		}
	}()

	ws.establishLoginConnection()

	ws.wg.Add(1)
	go ws.acceptNewServerConnections()

	ws.wg.Add(1)
	go ws.processEvent()

	ws.wg.Wait()
	log.Println("World Server stopped")
}

func (ws *worldServer) shutdown() {
	ws.cancel()
	// Stop accepting and unblock Accept()
	if ws.listener != nil {
		_ = ws.listener.Close()
	}
}

func (ws *worldServer) establishLoginConnection() {
	backoff := time.Second
	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
		}
		if ws.connectToLogin() {
			ws.state.RegisterWithLogin(ws.lconn)
			return
		}
		// Backoff with cap
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
			if backoff > 10*time.Second {
				backoff = 10 * time.Second
			}
		}
	}
}

func (ws *worldServer) connectToLogin() bool {
	dialAddr := ws.config.LoginAddress + ":" + ws.config.LoginPort
	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		log.Println("Could not connect to login server at", dialAddr)
		return false
	}

	log.Println("Connected to login server at", dialAddr)

	login := mnet.NewServer(conn, ws.eRecv, ws.config.PacketQueueSize)
	go login.Reader()
	go login.Writer()

	ws.lconn = login
	return true
}

func (ws *worldServer) acceptNewServerConnections() {
	defer ws.wg.Done()

	l, err := net.Listen("tcp", ws.config.ListenAddress+":"+ws.config.ListenPort)
	if err != nil {
		log.Println("world listen error:", err)
		// Fatal for serving new servers: stop the process loop
		ws.shutdown()
		return
	}
	ws.listener = l
	log.Println("Server listener ready:", ws.config.ListenAddress+":"+ws.config.ListenPort)
	defer func() { _ = l.Close() }()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ws.ctx.Err() != nil {
				// shutting down
				return
			}
			if isTempNetErr(err) {
				log.Println("temporary world Accept error:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("fatal world Accept error:", err)
			// Do not close ws.eRecv; stop only this accept loop.
			return
		}

		serverConn := mnet.NewServer(conn, ws.eRecv, ws.config.PacketQueueSize)
		go serverConn.Reader()
		go serverConn.Writer()
	}
}

func (ws *worldServer) processEvent() {
	defer ws.wg.Done()

	for {
		select {
		case <-ws.ctx.Done():
			log.Println("Stopping event handling: shutdown")
			return
		case e, ok := <-ws.eRecv:
			if !ok {
				log.Println("Stopping event handling: event channel closed")
				return
			}

			switch conn := e.Conn.(type) {
			case mnet.Server:
				switch e.Type {
				case mnet.MEServerConnected:
					log.Println("New server from", conn)
				case mnet.MEServerDisconnect:
					log.Println("Server at", conn, "disconnected")

					// If the login connection died, attempt to re-establish
					if conn == ws.lconn {
						log.Println("Attempting to re-establish login server connection")
						ws.establishLoginConnection()
					}

					ws.state.ServerDisconnected(conn)
				case mnet.MEServerPacket:
					ws.state.HandleServerPacket(conn, mpacket.NewReader(&e.Packet, time.Now().Unix()))
				}
			}
		}
	}
}
