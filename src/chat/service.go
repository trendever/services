package main

import (
	"chat/cmd"
	"os"
	//_ "chat/common"
	_ "chat/config"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"utils/log"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	})
}
