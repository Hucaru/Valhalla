package main

import (
	"flag"
	"log"

	"github.com/Hucaru/Valhalla/server"
)

func main() {
	typePtr := flag.String("type", "", "Denotes what type of server to start: login, world, channel")
	configPtr := flag.String("config", "config.toml", "config toml file")

	flag.Parse()

	switch *typePtr {
	case "login":
		s := server.NewLoginServer(*configPtr)
		s.Run()
	case "world":
		log.Println("World server not implemented yet")
	case "channel":
		server.Channel()
	default:
		log.Println("Unkown server type:", *typePtr)
	}
}
