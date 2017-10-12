package config

import (
	"common/config"
	"common/db"
	"common/log"
	"github.com/spf13/viper"
)

const (
	defaultPort      = "12325"
	defaultFromEmail = "hello@trendever.com"
)

type Settings struct {
	RPC   string
	From  string
	Debug bool
	DB    db.Settings

	MailCatcher struct {
		Addr string
	}
	MailGun struct {
		Domain       string
		APIKey       string
		PublicAPIKey string
	}

	SentryDSN string
}

var settings Settings

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("rpc", ":"+defaultPort)
	viper.SetDefault("debug", false)
	viper.SetDefault("mailcatcher.addr", "localhost:1025")
	viper.SetDefault("from", defaultFromEmail)

	if err := config.Load("mail"); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, "MAIL", settings.SentryDSN)
}

func Get() *Settings {
	return &settings
}
