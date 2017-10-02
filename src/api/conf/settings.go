package conf

import (
	"common/config"
	"common/log"
	"fmt"
	"github.com/spf13/viper"
	"utils/elastic"
	"utils/nats"
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
	Nats      nats.Config

	MarketSMS string

	API struct {
		Core         string
		Auth         string
		Chat         string
		Payments     string
		SMS          string
		Coins        string
		Checker      string
		Fetcher      string
		AccountStore string
	}

	Elastic elastic.Settings

	Mail struct {
		Username string
		Password string
		Host     string
		Sender   string
		Port     int
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

	PaymentsRedirects map[string]string
}
