package main

import (
	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
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

func loadDbConfig() *dbConfig {
	return &dbConfig{
		Address:  getEnv("DB_ADDRESS", ""),
		Port:     getEnv("DB_PORT", ""),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_NAME", ""),
	}
}

func loadLoginConfig() (loginConfig, dbConfig) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := &fullConfig{
		Database: *loadDbConfig(),
		Login: loginConfig{
			ClientListenAddress: getEnv("CLIENT_LISTEN_ADDRESS", ""),
			ClientListenPort:    getEnv("CLIENT_LISTEN_PORT", ""),
			ServerListenAddress: getEnv("LOGIN_LISTEN_ADDRESS", ""),
			ServerListenPort:    getEnv("LOGIN_LISTEN_PORT", ""),
			WithPin:             getEnvAsBool("WITH_PIN", false),
			PacketQueueSize:     getEnvAsInt("LOGIN_PACKET_QUEUE_SIZE", 0),
			Latency:             getEnvAsInt("LATENCY", 0),
			Jitter:              getEnvAsInt("JITTER", 0),
		},
	}

	return config.Login, config.Database
}

func loadWorldConfig() (worldConfig, dbConfig) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := &fullConfig{
		Database: *loadDbConfig(),
		World:    worldConfig{},
	}

	config.World.Message = getEnv("MESSAGE", "")
	config.World.Ribbon = getEnvAsByte("RIBBON", 2)
	config.World.ExpRate = getEnvAsFloat("EXP_RATE", 1.0)
	config.World.DropRate = getEnvAsFloat("DROP_RATE", 1.0)
	config.World.MesosRate = getEnvAsFloat("MESOS_RATE", 1.0)

	if getEnvAsBool("IS_DOCKER_CONFIG", true) {
		config.World.LoginAddress = getEnv("LOGIN_ADDRESS", "")
	} else {
		config.World.LoginAddress = getEnv("LOGIN_ADDRESS_LOCAL", "")
	}

	config.World.LoginPort = getEnv("LOGIN_LISTEN_PORT", "")
	config.World.ListenAddress = getEnv("WORLD_LISTEN_ADDRESS", "")
	config.World.ListenPort = getEnv("WORLD_LISTEN_PORT", "")
	config.World.PacketQueueSize = getEnvAsInt("WORLD_PACKET_QUEUE_SIZE", 0)

	return config.World, config.Database
}

func loadChannelConfig(fname string) (channelConfig, dbConfig) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := &fullConfig{}

	if _, err := toml.DecodeFile(fname, config); err != nil {
		log.Fatal(err)
	}
	log.Println("HELLP", config, fname)

	config.Database = *loadDbConfig()

	config.Channel.ListenAddress = getEnv("CHANNEL_LISTEN_ADDRESS", "")
	config.Channel.PacketQueueSize = getEnvAsInt("CHANNEL_PACKET_QUEUE_SIZE", 0)
	config.Channel.Jitter = getEnvAsInt("JITTER", 0)
	config.Channel.Latency = getEnvAsInt("LATENCY", 0)
	config.Channel.WorldPort = getEnv("WORLD_LISTEN_PORT", "")

	if getEnvAsBool("IS_DOCKER_CONFIG", true) {
		config.Channel.WorldAddress = getEnv("WORLD_ADDRESS", "")
	} else {
		config.Channel.WorldAddress = getEnv("WORLD_ADDRESS_LOCAL", "")
	}
	config.Channel.ClientConnectionAddress = getEnv("CLIENT_CONNECTION_ADDRESS", "")
	config.Channel.MaxPop = int16(getEnvAsInt("MAX_POP", 0))
	log.Println("DATA", config.Channel, config.Database)
	return config.Channel, config.Database
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnvAsFloat(name string, defaultVal float32) float32 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseFloat(valueStr, 32); err == nil {
		return float32(value)
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}

func getEnvAsByte(name string, defaultVal int) byte {
	valStr := getEnv(name, "")
	bs := make([]byte, getEnvAsInt(valStr, defaultVal))
	if len(bs) > 0 {
		return bs[0]
	}
	return 2
}
