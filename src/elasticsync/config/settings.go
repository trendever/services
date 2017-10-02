package config

import (
	"common/config"
	"common/db"
	"common/log"
	"fmt"
	"github.com/spf13/viper"
	"utils/elastic"
)

const (
	configName = "elasticsync"
	logTag     = "ELASTICSYNC"
)

//Settings is a app settings
type Settings struct {
	SentryDSN string
	NatsURL   string
	// delay between sync steps (msec)
	Delay int
	// maximum amount of documents that should be indexed on every step
	ChunkSize int
	Debug     bool

	DB      db.Settings
	Elastic elastic.Settings
}

var settings = &Settings{}

// Init loads config
func Init() {
	if err := config.Load(configName); err != nil {
		log.Fatal(fmt.Errorf("Can't load config: %v", err))
	}

	if err := viper.Unmarshal(settings); err != nil {
		log.Fatal(err)
	}

	log.Init(settings.Debug, logTag, settings.SentryDSN)
}

//Get returns an app settings
func Get() *Settings {
	return settings
}
