package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type dbConfig struct {
	Address  string
	Port     string
	User     string
	Password string
	Database string
}

type loginConfig struct {
	ClientListenAddress string
	ClientListenPort    string
	ServerListenAddress string
	ServerListenPort    string
	WithPin             bool
	PacketQueueSize     int
	Latency             int
	Jitter              int
}

type worldConfig struct {
	Message         string
	Ribbon          byte
	ExpRate         float32
	DropRate        float32
	MesosRate       float32
	LoginAddress    string
	LoginPort       string
	ListenAddress   string
	ListenPort      string
	PacketQueueSize int
}

type channelConfig struct {
	WorldAddress            string
	WorldPort               string
	ListenAddress           string
	ClientConnectionAddress string
	ListenPort              string
	PacketQueueSize         int
	MaxPop                  int16
	Latency                 int
	Jitter                  int
}

type fullConfig struct {
	Database dbConfig
	Login    loginConfig
	World    worldConfig
	Channel  channelConfig
}

func loginConfigFromFile(fname string) (loginConfig, dbConfig) {
	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}

	return config.Login, config.Database
}

func worldConfigFromFile(fname string) (worldConfig, dbConfig) {
	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}

	return config.World, config.Database
}

func channelConfigFromFile(fname string) (channelConfig, dbConfig) {
	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}

	return config.Channel, config.Database
}
