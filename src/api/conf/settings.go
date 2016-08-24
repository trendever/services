package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/elastic"
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

	viper.SetDefault("CoreAddr", "core:3005")
	viper.SetDefault("AuthAddr", "auth:8433")
	viper.SetDefault("ChatAddr", "chat:2010")
	viper.SetDefault("PaymentsAddr", "payments:7777")

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
	Debug       bool
	ChannelPort string

	SentryDSN string
	NatsURL   string

	CoreAddr     string
	AuthAddr     string
	ChatAddr     string
	PaymentsAddr string

	Elastic elastic.Settings

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
		DB       int
	}
}
