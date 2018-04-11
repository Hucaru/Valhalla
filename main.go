package main

import (
	"flag"
	"log"

	"github.com/Hucaru/Valhalla/server"
)

func main() {
	typePtr := flag.String("type", "", "Denotes what type of server to start: login, world, channel")

	flag.Parse()

	switch *typePtr {
	case "login":
		server.Login()
	case "world":
		log.Println("World server not implemented yet")
	case "channel":
		server.Channel()
	default:
		log.Println("Unkown server type:", *typePtr)
	}

}
