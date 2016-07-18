package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/log"
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
}

// GetSettings returns service settings
func GetSettings() *Settings {
	return settings
}

type Settings struct {
	Debug     bool
	RPC       string
	DB        DB
	Instagram Instagram
	SentryDSN string
}

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	Debug    bool
}

type Instagram struct {
	TimeoutMin int
	TimeoutMax int
	Users      []struct {
		Username string
		Password string
	} `yaml:"users"`
}
