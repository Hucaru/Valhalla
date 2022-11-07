package main

import (
	"flag"
	"log"
	"sync"

	"github.com/Hucaru/Valhalla/common"
	"github.com/spf13/pflag"
)

var (
	once      sync.Once
	typePtr   = pflag.String("type", "channel", "Denotes what type of server to start: login, world, channel")
	configPtr = pflag.String("config", "config_channel_1.toml", "config toml file")
	metricPtr = pflag.String("metrics-port", "", "Port to serve metrics on")
)

func init() {
	once.Do(func() {
		log.Println("Parsing flags")
		parseFlags()
	})
}

func main() {
	common.MetricsPort = *metricPtr
	//
	log.Println("TYPE", *typePtr)

	s := newChannelServer(*configPtr)
	s.run()

	//switch *typePtr {
	//case "login":
	//	s := newLoginServer(*configPtr)
	//	s.run()
	//case "world":
	//	s := newWorldServer(*configPtr)
	//	s.run()
	//case "channel":
	//	s := newChannelServer(*configPtr)
	//	s.run()
	//default:
	//	log.Println("Unknown server type:", typePtr)
	//}
}

func parseFlags() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	//_ = flag.Lookup("logtostderr").Value.Set("true")
}
