package main

import (
	"log"

	"github.com/Hucaru/Valhalla/worldServer/handlers"
	"github.com/Hucaru/Valhalla/worldServer/handlers/channel"
	"github.com/Hucaru/Valhalla/worldServer/handlers/login"
)

func main() {
	log.Println("WorldServer")

	go login.Connect()
	go channel.Handle()

	handlers.Run()
}
