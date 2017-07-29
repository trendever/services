package config

import (
	"chat/notifier"
	"github.com/spf13/viper"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/nats"
)

const consulConfigPath = "chat"

//Settings is app config structure
type Settings struct {
	Port          string
	Host          string
	Debug         bool
	DB            db.Settings
	Receiver      string
	UploadService string `mapstructure:"upload_service"`
	SentryDSN     string
	RPC           struct {
		Core    string
		Checker string
	}
	Nats       nats.Config
	Unanswered map[string]notifier.Config
}

var settings = &Settings{}

// Init loads && setups config
func Init() {
	if err := config.Load(consulConfigPath); err != nil {
		log.Error(err)
	}

	viper.SetDefault("Port", "2010")
	viper.SetDefault("Host", "localhost")
	viper.SetDefault("upload_service", "http://localhost:8080")
	viper.SetDefault("Debug", true)
	viper.SetDefault("Receiver", ":2011,:2012")
	if err := viper.Unmarshal(settings); err != nil {
		log.Fatal(err)
	}
	log.Init(settings.Debug, "CHAT", settings.SentryDSN)
}

//Get returns app settings
func Get() *Settings {
	return settings
}
