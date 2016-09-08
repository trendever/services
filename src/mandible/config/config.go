package config

import (
	"log"

	"utils/config"
)

type Configuration struct {
	MaxFileSize     int64
	HashLength      int
	UserAgent       string
	Stores          []map[string]string
	Port            int
	DatadogEnabled  bool
	DatadogHostname string
}

const configName = "mandible"

func NewConfiguration() *Configuration {
	configuration := Configuration{}
	err := config.LoadStruct(configName, &configuration)

	if err != nil {
		log.Fatal(err)
	}

	return &configuration
}
