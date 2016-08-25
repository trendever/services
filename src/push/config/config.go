package config

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/db"
	"utils/log"
)

const (
	configName = "push"
	logTag     = "PUSH"
)

//Settings is a app settings
type Settings struct {
	RPC              string
	PushTokensServer string
	FMCServerKey     string
	APNPemFile       string
	APNPemPass       string
	APNTopic         string
	DB               db.Settings

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
