package server

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
	PacketQueueSize     int
}

type worldConfig struct {
	Message         string
	Ribbon          byte
	LoginAddress    string
	LoginPort       string
	ListenAddress   string
	ListenPort      string
	PacketQueueSize int
}

type channelConfig struct {
	WorldAddress    string
	WorldPort       string
	ListenAddress   string
	ListenPort      string
	PacketQueueSize int
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
