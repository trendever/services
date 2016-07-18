package config

import (
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
)

const consulConfigPath = "chat"

//Settings is app config structure
type Settings struct {
	Port          string
	Host          string
	Debug         bool
	DB            struct{ Conf string }
	Receiver      string
	UploadService string `mapstructure:"upload_service"`
	SentryDSN     string
	NatsURL       string
}

var settings = &Settings{}

func init() {
	if err := config.Load(consulConfigPath); err != nil {
		log.Error(err)
	}

	viper.SetDefault("Port", "2010")
	viper.SetDefault("Host", "localhost")
	viper.SetDefault("upload_service", "http://localhost:8080")
	viper.SetDefault("DB", struct{ Conf string }{Conf: "user=postgres dbname=postgres password=1234 sslmode=disable"})
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
