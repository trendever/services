package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	// needed for remote configs
	_ "github.com/spf13/viper/remote"
)

const (
	configFormat = "yaml"
)

// Load loads config viper config
// if env variable CONSUL_HOST is set, it will try to get config with key name remoteName from it
// by default following paths are searched:
//   $WEB_ROOT/
//   $PWD/
//   $WEB_ROOT/config
//   $PWD/config
// for the file named 'config.yaml'
func Load(name string) error {
	viper.SetConfigName(name)
	viper.SetConfigType(configFormat)

	viper.AddConfigPath("$WEB_ROOT/")
	viper.AddConfigPath(".")

	viper.AddConfigPath("$WEB_ROOT/config")
	viper.AddConfigPath("./config")

	if remoteHost := os.Getenv("ETCD_HOST"); remoteHost != "" {

		if configPath := os.Getenv("ETCD_PATH"); configPath != "" {
			name = configPath
		}

		// read remote
		viper.AddRemoteProvider("etcd", remoteHost, fmt.Sprintf("%v.%v", name, configFormat))

		return viper.ReadRemoteConfig()
	}

	// read local
	return viper.ReadInConfig()
}
