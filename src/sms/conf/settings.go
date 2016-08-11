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
const ()

var (
	debugTag         = "SMS"
	consulConfigPath = "sms"
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

	log.Init(settings.Debug, debugTag, settings.SentryDSN)
}

//GetSettings returns Settings object
func GetSettings() *Settings {
	return settings
}

//Settings is program settings
type Settings struct {
	Debug     bool
	Rpc       string
	DB        DB
	Sender    string
	Atompark  Atompark
	SentryDSN string
	MTS       struct {
		Login    string
		Password string
		Naming   string
		Rates    int
	}
	Telegram struct {
		Rpc     string
		Channel string
	}
}

//DB is database settings
type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	Debug    bool
}

//Atompark is Atompark settings
type Atompark struct {
	Test       string
	KeyPublic  string
	KeyPrivate string
	Sender     string
}
