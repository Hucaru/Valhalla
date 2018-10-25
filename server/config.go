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
	ListenAddress   string
	ListenPort      string
	PacketQueueSize int
}

type channelConfig struct {
	ListenAddress   string
	ListenPort      string
	PacketQueueSize int
}

type fullConfig struct {
	Database dbConfig
	Login    loginConfig
	Channel  channelConfig
}

func loginConfigFromFile(fname string) (loginConfig, dbConfig) {
	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}

	return config.Login, config.Database
}

func channelConfigFromFile(fname string) (channelConfig, dbConfig) {
	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}

	return config.Channel, config.Database
}
