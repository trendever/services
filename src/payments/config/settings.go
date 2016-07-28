package config

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
)

const configName = "payment"

//Settings is a app settings
type Settings struct {
	RPC          string
	CallbackHTTP string
	Debug        bool
	SentryDSN    string
	DB           struct {
		Conf  string
		Debug bool
	}
	Payture struct {
		Sandbox bool
		URL     string
		Key     string
	}
}

var settings = &Settings{}

func init() {
	viper.SetDefault("rpc", ":7777")

	if err := config.Load(configName); err != nil {
		log.Fatal(fmt.Errorf("Can't load config: %v", err))
	}

	if err := viper.Unmarshal(settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, "PAYMENT", settings.SentryDSN)
}

//Get returns an app settings
func Get() Settings {
	return *settings
}
