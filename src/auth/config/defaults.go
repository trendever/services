package config

import (
	"errors"
	"github.com/dvsekhvalnov/jose2go"
	"github.com/spf13/viper"
	"utils/config"
	"utils/db"
	"utils/log"
)

const (
	defaultPort = "12443"
	defaultAlg  = jose.HS256
)

type Config struct {
	Port        string
	Host        string
	CoreServer  string `mapstructure:"core_server"`
	SmsServer   string `mapstructure:"sms_server"`
	PasswordLen int    `mapstructure:"password_len"`
	Key         string
	Debug       bool
	DB          db.Settings
	NatsURL     string
	SiteUrl     string `mapstructure:"site_url"`
	SmsTemplate string `mapstructure:"sms_template"`
	SentryDSN   string
}

var settings = &Config{}

func init() {

	viper.SetDefault("port", defaultPort)
	viper.SetDefault("alg", defaultAlg)
	viper.SetDefault("password_len", 6)
	if err := config.Load("auth"); err != nil {
		log.Fatal(err)
	}

	if err := viper.Unmarshal(&settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, "AUTH", settings.SentryDSN)

	if settings.SmsTemplate == "" {
		log.Fatal(errors.New("sms_template is empty! Check your config"))
	}
	if settings.SiteUrl == "" {
		log.Fatal(errors.New("site_url is empty! Check your config"))
	}

}

func Get() *Config {
	return settings
}
