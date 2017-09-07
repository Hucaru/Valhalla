package main

import (
	"github.com/Hucaru/Valhalla/loginServer/handlers"
	"github.com/Hucaru/Valhalla/loginServer/handlers/channel"
	"github.com/Hucaru/Valhalla/loginServer/handlers/client"
	"github.com/Hucaru/Valhalla/loginServer/handlers/world"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	go world.StartListening()
	go channel.StartListening()
	go handlers.Manager()

	client.Handle()
}
