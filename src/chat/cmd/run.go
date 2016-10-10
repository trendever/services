package cmd

import (
	"chat/config"
	"chat/models"
	"chat/queue"
	"chat/server"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/chat"
	"syscall"
	"time"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		conf := config.Get()

		log.Info("Starting chat service on %s:%s", conf.Host, conf.Port)
		s := rpc.Serve(conf.Host + ":" + conf.Port)
		db.Init(&conf.DB)
		db := db.New()

		repository := models.NewConversationRepository(db)
		nats.Init(conf.NatsURL)

		chat.RegisterChatServiceServer(s, server.NewChatServer(
			repository,
			queue.NewWaiter(time.Minute*5),
		))

		// interrupt
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		// wait for terminating
		for {
			select {
			case <-interrupt:
				log.Info("Service stopped")
				os.Exit(0)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cmdRun)
}
