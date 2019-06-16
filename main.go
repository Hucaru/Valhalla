package main

import (
	"flag"
	"log"

	"github.com/Hucaru/Valhalla/server"
)

func main() {
	typePtr := flag.String("type", "", "Denotes what type of server to start: login, world, channel")
	configPtr := flag.String("config", "", "config toml file")

	flag.Parse()

	switch *typePtr {
	case "login":
		s := server.NewLoginServer(*configPtr)
		s.Run()
	case "world":
		s := server.NewWorldServer(*configPtr)
		s.Run()
	case "channel":
		s := server.NewChannelServer(*configPtr)
		s.Run()
	default:
		log.Println("Unkown server type:", *typePtr)
	}
}
