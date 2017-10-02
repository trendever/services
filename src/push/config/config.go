package config

import (
	"common/config"
	"common/db"
	"common/log"
	"fmt"
	"github.com/spf13/viper"
)

const (
	configName = "push"
	logTag     = "PUSH"
)

//Settings is a app settings
type Settings struct {
	RPC              string
	PushTokensServer string
	// minimal timeout between attempts to send message to receiver
	RetryTimeout uint64

	FCMServerKey string
	APNPemFile   string
	APNPemPass   string
	APNSandbox   bool
	APNTopic     string
	DB           db.Settings

	Debug     bool
	SentryDSN string
}

var settings = &Settings{}

// Init loads config
func Init() {
	if err := config.Load(configName); err != nil {
		log.Fatal(fmt.Errorf("Can't load config: %v", err))
	}

	if err := viper.Unmarshal(settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, logTag, settings.SentryDSN)
}

//Get returns an app settings
func Get() *Settings {
	return settings
}
