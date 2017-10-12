package main

import (
	"auth/cmd"
	_ "auth/config"
	"common/log"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
