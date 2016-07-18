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
	LastCheckedFile  = "last.txt"
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
	Debug         bool
	CoreServer    string
	FetcherServer string
	Instagram     Instagram
	SentryDSN     string
	MandibleURL   string
}

type Instagram struct {
	TimeoutMin     int
	TimeoutMax     int
	ReloginTimeout int
	PollTimeout    int
	TrendUser      string
	Users          []struct {
		Username string
		Password string
	}
}
