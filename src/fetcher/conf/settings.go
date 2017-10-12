package conf

import (
	"common/config"
	"common/db"
	"common/log"
	"fmt"
	"instagram"

	"accountstore/client"
	"github.com/spf13/viper"
	"utils/nats"
)

var (
	settings *Settings
)

// These are settings, but without need to change them in a file
const (
	consulConfigPath = "fetcher"
	tagName          = "Fetcher"
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
	instagram.DisableJSONIndent = !settings.Instagram.IndentLogging
}

// GetSettings returns service settings
func GetSettings() *Settings {
	return settings
}

// Settings main config
type Settings struct {
	Debug     bool
	RPC       string
	DB        db.Settings
	Instagram Instagram
	SentryDSN string
	Nats      nats.Config
}

// Instagram config part
type Instagram struct {
	client.Settings `mapstructure:",squash"`
	StoreAddr       string
	ResponseLogging bool
	IndentLogging   bool
}
