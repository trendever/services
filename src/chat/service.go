package main

import (
	"os"
	"chat/cmd"
	//_ "chat/common"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"utils/log"
	_ "chat/config"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	})
}
