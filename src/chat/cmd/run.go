package cmd

import (
	"chat/config"
	"chat/db"
	"chat/models"
	"chat/publisher"
	"chat/queue"
	"chat/server"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/chat"
	"syscall"
	"time"
	"utils/log"
	"utils/rpc"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.Get()

		log.Info("Starting chat service on %s:%s", conf.Host, conf.Port)
		s := rpc.Serve(conf.Host + ":" + conf.Port)
		db := db.GetPG()
		repository := models.NewConversationRepository(db)
		publisher.Init(repository)
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
