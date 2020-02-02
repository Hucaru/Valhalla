package main

import (
	"flag"
	"log"
)

func main() {
	typePtr := flag.String("type", "", "Denotes what type of server to start: login, world, channel")
	configPtr := flag.String("config", "", "config toml file")

	flag.Parse()

	switch *typePtr {
	case "login":
		s := newLoginServer(*configPtr)
		s.run()
	case "world":
		s := newWorldServer(*configPtr)
		s.run()
	case "channel":
		s := newChannelServer(*configPtr)
		s.run()
	default:
		log.Println("Unkown server type:", *typePtr)
	}
}
