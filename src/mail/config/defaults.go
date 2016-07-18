package config

import (
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
)

const (
	defaultDbDriver  = "postgres"
	defaultDbConfig  = "user=postgres dbname=postgres password=1234 sslmode=disable"
	defaultPort      = "12325"
	defaultFromEmail = "hello@trendever.com"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("port", defaultPort)
	viper.SetDefault("db.dirver", defaultDbDriver)
	viper.SetDefault("db.config", defaultDbConfig)
	viper.SetDefault("debug", false)
	viper.SetDefault("mailcatcher.addr", "localhost:1025")
	viper.SetDefault("from", defaultFromEmail)

	if err := config.Load("mail"); err != nil {
		log.Fatal(err)
	}

	log.Init(viper.GetBool("debug"), "MAIL", viper.GetString("sentryDSN"))

}
