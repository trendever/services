package main

import (
	_ "github.com/lib/pq"
	"utils/log"
	"os"
	"sms/cmd"
)

func main() {
	log.PanicLogger(func() {
		if err := cmd.RootCmd.Execute(); err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	})
}
