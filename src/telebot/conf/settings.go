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
	consulConfigPath = "telebot"
	tagName          = "Telegram"
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

// Settings defines service configuration
type Settings struct {
	Debug bool
	RPC   string
	Token string
	Rooms []Room
	SentryDSN string
}

// Room defines telegram chat room (common name and real @room)
type Room struct {
	Name string
	Room string
}
