package main

import (
	"mail/cmd"
	_ "mail/common"
	_ "mail/config"
	"os"
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
