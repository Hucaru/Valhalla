package main

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type dbConfig struct {
	Address  string	`mapstructure:"address"`
	Port     string	`mapstructure:"port"`
	User     string	`mapstructure:"user"`
	Password string	`mapstructure:"password"`
	Database string	`mapstructure:"database"`
}

type loginConfig struct {
	ClientListenAddress string	`mapstructure:"clientListenAddress"`
	ClientListenPort    string	`mapstructure:"clientListenPort"`
	ServerListenAddress string	`mapstructure:"serverListenAddress"`
	ServerListenPort    string	`mapstructure:"serverListenPort"`
	WithPin             bool	`mapstructure:"withPin"`
	AutoRegister        bool	`mapstructure:"autoRegister"`
	PacketQueueSize     int		`mapstructure:"packetQueueSize"`
	Latency             int		`mapstructure:"latency"`
	Jitter              int		`mapstructure:"jitter"`
}

type worldConfig struct {
	Message         string		`mapstructure:"message"`
	Ribbon          byte		`mapstructure:"ribbon"`
	ExpRate         float32		`mapstructure:"expRate"`
	DropRate        float32		`mapstructure:"dropRate"`
	MesosRate       float32		`mapstructure:"mesosRate"`
	LoginAddress    string		`mapstructure:"loginAddress"`
	LoginPort       string		`mapstructure:"loginPort"`
	ListenAddress   string		`mapstructure:"listenAddress"`
	ListenPort      string		`mapstructure:"listenPort"`
	PacketQueueSize int			`mapstructure:"packetQueueSize"`
}

type channelConfig struct {
	WorldAddress            string	`mapstructure:"worldAddress"`
	WorldPort               string	`mapstructure:"worldPort"`
	ListenAddress           string	`mapstructure:"listenAddress"`
	ClientConnectionAddress string	`mapstructure:"clientConnectionAddress"`
	ListenPort              string	`mapstructure:"listenPort"`
	PacketQueueSize         int		`mapstructure:"packetQueueSize"`
	MaxPop                  int16	`mapstructure:"maxPop"`
	Latency                 int		`mapstructure:"latency"`
	Jitter                  int		`mapstructure:"jitter"`
}

type cashShopConfig struct {
	WorldAddress            string	`mapstructure:"worldAddress"`
	WorldPort               string	`mapstructure:"worldPort"`
	ListenAddress           string	`mapstructure:"listenAddress"`
	ClientConnectionAddress string	`mapstructure:"clientConnectionAddress"`
	ListenPort              string	`mapstructure:"listenPort"`
	PacketQueueSize         int		`mapstructure:"packetQueueSize"`
	Latency                 int		`mapstructure:"latency"`
	Jitter                  int		`mapstructure:"jitter"`
}

type fullConfig struct {
	Database dbConfig		`mapstructure:"database"`
	Login    loginConfig	`mapstructure:"login"`
	World    worldConfig	`mapstructure:"world"`
	Channel  channelConfig	`mapstructure:"channel"`
	CashShop cashShopConfig	`mapstructure:"cashshop"`
}

// Load from TOML if exists, then load/overwrite with ENV
func LoadConfig(fname string) *fullConfig {
	v := viper.New()

	if fname != "" {
		v.SetConfigFile(fname)
		v.SetConfigType("toml")
		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if errors.As(err, &notFound) {
				log.Printf("warning: config file %q not found; continuing with env", fname)
			} else {
				log.Fatalf("failed to read config file %q: %v", fname, err)
			}
		}
	}

	v.SetEnvPrefix("VALHALLA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	bindEnvs(v, reflect.TypeOf(fullConfig{}), nil, "VALHALLA")

	var config fullConfig
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &config
}

func loginConfigFromFile(fname string) (loginConfig, dbConfig) {
	config := LoadConfig(fname)
	return config.Login, config.Database
}

func worldConfigFromFile(fname string) (worldConfig, dbConfig) {
	config := LoadConfig(fname)
	return config.World, config.Database
}

func channelConfigFromFile(fname string) (channelConfig, dbConfig) {
	config := LoadConfig(fname)
	return config.Channel, config.Database
}

func cashShopConfigFromFile(fname string) (cashShopConfig, dbConfig) {
	config := LoadConfig(fname)
	return config.CashShop, config.Database
}

func bindEnvs(v *viper.Viper, typ reflect.Type, path []string, envPrefix string) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}

		tag := f.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue // skip untagged
		}

		fullPath := append(path, tag)
		if f.Type.Kind() == reflect.Struct {
			bindEnvs(v, f.Type, fullPath, envPrefix)
			continue
		}
		cfgKey := strings.Join(fullPath, ".")
		env := envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(cfgKey, ".", "_"))
		_ = v.BindEnv(cfgKey, env)
	}
}
