package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"payments/cmd"
	"utils/log"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
