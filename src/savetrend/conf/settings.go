package conf

import (
	"accountstore/client"
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
	"utils/nats"
)

var (
	settings *Settings
)

// These are settings, but without need to change them in a file
const (
	consulConfigPath = "savetrend"
	tagName          = "Savetrend"
)

func init() {

	err := config.Load(consulConfigPath)

	if err != nil {
		panic(fmt.Errorf("Config not loaded: %v", err))
	}

	err = viper.Unmarshal(&settings)
	if err != nil {
		panic(fmt.Errorf("Config can not be unmarshaled: %v", err))
	}

	log.Init(settings.Debug, tagName, settings.SentryDSN)
}

func GetSettings() *Settings {
	return settings
}

type Settings struct {
	Debug                  bool
	Rpc                    string
	CoreServer             string
	FetcherServer          string
	Instagram              Instagram
	SentryDSN              string
	MandibleURL            string
	LastCheckedFile        string
	DirectNotificationText string
	Nats                   nats.Config
}

// Instagram config
type Instagram struct {
	client.Settings `mapstructure:",squash"`
	StoreAddr       string
	ResponseLogging bool
}
