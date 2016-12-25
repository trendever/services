package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/nats"
)

var (
	settings *Settings
)

// These are settings, but without need to change them in a file
const (
	AdminName         = "Trendever"
	configName        = "core"
	defaultSiteURL    = "https://trendever.com"
	defaultSystemUser = "trendever"
)

// Init loads configuration
func Init() {

	viper.SetDefault("SiteURL", defaultSiteURL)

	err := config.Load(configName)

	if err != nil {
		panic(fmt.Errorf("Config not loaded: %v", err))
	}

	err = viper.Unmarshal(&settings)
	if err != nil {
		panic(fmt.Errorf("Config can not be unmarshaled: %v", err))
	}

	log.Init(settings.Debug, "CORE", settings.SentryDSN)
	if settings.SystemUser == "" {
		settings.SystemUser = defaultSystemUser
	}
}

// GetSettings returns current settings instance
func GetSettings() *Settings {
	return settings
}

// Settings container
type Settings struct {
	Debug      bool
	AppHost    string
	SiteURL    string
	SentryDSN  string
	SystemUser string

	Monetization struct {
		// name of plan that will be assigned to shop after creation. default is 'init'
		InitialPlanName string
		// duration multiplier for plans periods
		PlansBaseDuration     string
		SubscriptionCheckTick string
	}

	Comments struct {
		Allowed string
	}

	RPC struct {
		Listen    string
		Mail      string
		SMS       string
		Chat      string
		Auth      string
		Push      string
		Telegram  string
		Checker   string
		Trendcoin string
		Payments  string
	}

	DB   db.Settings
	Nats nats.Config

	Bitly struct {
		APIKey      string
		Login       string
		AccessToken string
	}
	Profiler struct {
		Web  bool
		Addr string
	}
}
