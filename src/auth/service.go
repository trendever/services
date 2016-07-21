package main

import (
	"auth/cmd"
	_ "auth/config"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"utils/log"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
