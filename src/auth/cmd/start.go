package cmd

import (
	"auth/db"
	"auth/models"
	"auth/server"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	protocol "proto/auth"
	protocol_core "proto/core"
	protocol_sms "proto/sms"
	"syscall"
	"utils/log"
	"utils/rpc"
)

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := db.GetPG(viper.GetString("db.config"))
		if err != nil {
			log.Fatal(err)
		}
		port := viper.GetString("port")
		host := viper.GetString("host")
		log.Info("Starting auth microservice on the port: %s:%s", host, port)
		s := rpc.Serve(fmt.Sprintf("%s:%s", host, port))
		sms := protocol_sms.NewSmsServiceClient(rpc.Connect(viper.GetString("sms_server")))
		core := protocol_core.NewUserServiceClient(rpc.Connect(viper.GetString("core_server")))
		key := []byte(viper.GetString("key"))

		if len(key) < 16 {
			panic(fmt.Errorf("Bad key (key len should be at least 16 bytes, got %v bytes)", len(key)))
		}

		protocol.RegisterAuthServiceServer(s, server.NewAuthServer(core, sms, models.MakeNewUserPasswords(db), viper.GetString("alg"), key))
		// interrupt
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		// wait for terminating
		for {
			select {
			case <-interrupt:
				log.Warn("Cleanup and terminating...")
				os.Exit(0)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cmdStart)
}
