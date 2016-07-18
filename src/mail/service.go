package main

import (
	"utils/log"
	"os"
	"mail/cmd"
	_ "mail/common"
	_ "mail/config"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	})
}
