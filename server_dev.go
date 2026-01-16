package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Hucaru/Valhalla/common"
)

type devServer struct {
	configFile   string
	wg           *sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	loginServer  *loginServer
	worldServer  *worldServer
	channelServer *channelServer
	cashShopServer *cashShopServer
}

func newDevServer(configFile string) *devServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &devServer{
		configFile: configFile,
		wg:         &sync.WaitGroup{},
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (ds *devServer) run() {
	log.Println("===============================================")
	log.Println("Dev Server - All-in-One Mode")
	log.Println("Starting Login, World, Channel, and CashShop servers")
	log.Println("===============================================")

	// Signal handler for graceful shutdown
	ds.wg.Add(1)
	go func() {
		defer ds.wg.Done()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigCh:
			log.Println("Shutdown signal received")
			ds.shutdown()
		case <-ds.ctx.Done():
		}
	}()

	// Start all servers in goroutines
	// Login server starts first
	ds.wg.Add(1)
	go func() {
		defer ds.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println("Login server panic:", r)
			}
		}()
		ds.loginServer = newLoginServer(ds.configFile)
		ds.loginServer.run()
	}()

	// Give login server time to start
	time.Sleep(2 * time.Second)

	// Start world server
	ds.wg.Add(1)
	go func() {
		defer ds.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println("World server panic:", r)
			}
		}()
		ds.worldServer = newWorldServer(ds.configFile)
		ds.worldServer.run()
	}()

	// Give world server time to connect to login
	time.Sleep(2 * time.Second)

	// Start channel server
	ds.wg.Add(1)
	go func() {
		defer ds.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println("Channel server panic:", r)
			}
		}()
		ds.channelServer = newChannelServer(ds.configFile)
		ds.channelServer.run()
	}()

	// Give channel server time to connect to world
	time.Sleep(2 * time.Second)

	// Start cashshop server
	ds.wg.Add(1)
	go func() {
		defer ds.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println("CashShop server panic:", r)
			}
		}()
		ds.cashShopServer = newCashShopServer(ds.configFile)
		ds.cashShopServer.run()
	}()

	log.Println("===============================================")
	log.Println("All servers started successfully!")
	log.Println("Connect to: 127.0.0.1:8484")
	log.Println("Press Ctrl+C to stop all servers")
	log.Println("===============================================")

	// Wait for all servers to exit
	ds.wg.Wait()
	log.Println("Dev Server stopped")
}

func (ds *devServer) shutdown() {
	log.Println("Shutting down all servers...")
	
	// Trigger shutdown on all servers
	// Note: Each server has its own signal handler that will also trigger
	// on SIGINT/SIGTERM, so they will shutdown gracefully on their own.
	// This is a backup to ensure shutdown is triggered.
	if ds.cashShopServer != nil {
		ds.cashShopServer.shutdown()
	}
	if ds.channelServer != nil {
		ds.channelServer.shutdown()
	}
	if ds.worldServer != nil {
		ds.worldServer.shutdown()
	}
	if ds.loginServer != nil {
		ds.loginServer.shutdown()
	}
	
	// Stop the metrics server
	common.StopMetrics()
	
	ds.cancel()
}
