package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
)

var settings *Settings

// These are settings, but without need to change them in a file
const (
	LogTag           = "API"
	consulConfigPath = "api"
)

func init() {

	err := config.Load(consulConfigPath)

	viper.SetDefault("Debug", true)
	viper.SetDefault("ChannelPort", 8081)
	viper.SetDefault("CoreAddr", "localhost:3005")
	viper.SetDefault("AuthAddr", "localhost:8433")
	viper.SetDefault("ChatAddr", "localhost:2010")

	if err != nil {
		panic(fmt.Errorf("Config not loaded: %v", err))
	}

	err = viper.Unmarshal(&settings)
	if err != nil {
		panic(fmt.Errorf("Config can not be unmarshaled: %v", err))
	}

	log.Init(settings.Debug, LogTag, settings.SentryDSN)
}

func GetSettings() *Settings {
	return settings
}

type Settings struct {
	Debug       bool   `default:"true"`
	ChannelPort string `default:"8081"`
	CoreAddr    string `default:"localhost:3005"`
	AuthAddr    string `default:"localhost:8433"`
	ChatAddr    string
	SentryDSN   string
	NatsURL     string

	Mail struct {
		Username string
		Password string
		Host     string
		Sender   string
		Port     int
	}

	DB struct {
		User     string
		Password string
		Host     string
		Port     string
		Name     string
	}

	Metrics struct {
		Addr     string
		User     string
		Password string
		DBName   string
	}

	Profiler struct {
		Web  bool
		Addr string
	}

	Redis struct {
		Addr     string
		Password string
		DB       int64
	}
}
