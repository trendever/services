package config

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/nats"
)

const configName = "payment"

//Settings is a app settings
type Settings struct {
	RPC         string
	ChatServer  string
	CoinsServer string

	Debug     bool
	SentryDSN string
	DB        db.Settings
	Nats      nats.Config
	Payture   struct {
		Sandbox bool
		URL     string
		Key     string
	}
	HTTP struct {
		Listen   string // http-server bind addr (like :7780)
		Public   string // public-accessible URL of http-server root (like http://te.com:7780/)
		Redirect string // success redirect URL. Format string: 1st %v -- success bool; 2nd -- lead id (may be zero)
	}

	PeriodicCheck int // how often do we check unfinished txs; secs
}

var settings = &Settings{}

// Init loads config
func Init() {
	viper.SetDefault("rpc", ":7777")
	viper.SetDefault("http.listen", ":7780")
	viper.SetDefault("periodicCheck", "120")
	viper.SetDefault("chatServer", "chat:2010")

	if err := config.Load(configName); err != nil {
		log.Fatal(fmt.Errorf("Can't load config: %v", err))
	}

	if err := viper.Unmarshal(settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, "PAYMENT", settings.SentryDSN)
}

//Get returns an app settings
func Get() *Settings {
	return settings
}
