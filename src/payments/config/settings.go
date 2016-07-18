package config

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
)

const consulConfigPath = "payment"

//Settings is a app settings
type Settings struct {
	Listen    string
	Debug     bool
	SentryDSN string
	DB        struct {
		Conf  string
		Debug bool
	}
}

var settings = &Settings{}

func init() {
	viper.SetDefault("listen", ":7777")

	if err := config.Load(consulConfigPath); err != nil {
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
