package main

import (
	"chat/cmd"
	"os"
	//_ "chat/common"
	_ "chat/config"
	"common/log"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	})
}
