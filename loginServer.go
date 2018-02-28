package main

import (
	"github.com/Hucaru/Valhalla/handlers"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	handlers.Login()
}
