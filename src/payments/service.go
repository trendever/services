package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"utils/log"
	"payments/cmd"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
