package main

import (
	"flag"
	"log"

	"github.com/Hucaru/Valhalla/server"
)

func main() {
	typePtr := flag.String("type", "", "Denotes what type of server to start: login, world, channel")
	configPtr := flag.String("config", "", "Config file to use with server")

	flag.Parse()

	switch *typePtr {
	case "login":
		server.Login(*configPtr)
	case "world":
		log.Println("World server not implemented yet")
	case "channel":
		server.Channel(*configPtr)
	default:
		log.Println("Unkown server type:", *typePtr)
	}

}
