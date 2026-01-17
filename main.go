package main

import (
	"flag"
	"log"

	"github.com/Hucaru/Valhalla/common"
)

var typePtr, configPtr, metricPtr *string
var channelPtr *int

func init() {
	typePtr = flag.String("type", "", "Denotes what type of server to start: login, world, channel, cashshop, dev")
	configPtr = flag.String("config", "", "config toml file")
	metricPtr = flag.String("metrics-port", "9000", "Port to serve metrics on")
	channelPtr = flag.Int("channels", 2, "Defines number of channels to start (only for dev server type)")
	flag.Parse()
}

func main() {
	common.MetricsPort = *metricPtr

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
	case "cashshop":
		s := newCashShopServer(*configPtr)
		s.run()
	case "dev":
		s := newDevServer(*configPtr)
		s.run()
	default:
		log.Println("Unknown server type:", *typePtr)
	}
}
