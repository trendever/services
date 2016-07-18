package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"utils/log"
	"auth/cmd"
	_ "auth/config"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
