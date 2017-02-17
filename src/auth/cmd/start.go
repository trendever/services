package cmd

import (
	"auth/config"
	"auth/models"
	"auth/server"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	protocol "proto/auth"
	protocol_core "proto/core"
	protocol_sms "proto/sms"
	"syscall"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.Get()
		db.Init(&conf.DB)
		nats.Init(&conf.Nats, true)

		addr := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
		log.Info("Starting auth microservice on %s", addr)
		s := rpc.Serve(addr)
		sms := protocol_sms.NewSmsServiceClient(rpc.Connect(conf.SmsServer))
		core := protocol_core.NewUserServiceClient(rpc.Connect(conf.CoreServer))
		key := []byte(conf.Key)

		if len(key) < 16 {
			log.Fatal(fmt.Errorf("Bad key (key len should be at least 16 bytes, got %v bytes)", len(key)))
		}

		protocol.RegisterAuthServiceServer(s, server.NewAuthServer(core, sms, models.MakeNewUserPasswords(db.New()), key))

		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		<-interrupt
		log.Warn("Cleanup and terminating...")
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(cmdStart)
}
