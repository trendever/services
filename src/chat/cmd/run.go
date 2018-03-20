package cmd

import (
	"chat/config"
	"chat/models"
	"chat/notifier"
	"chat/server"
	"common/db"
	"common/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/chat"
	"syscall"
	"utils/nats"
	"utils/rpc"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		conf := config.Get()

		models.InitUploader(conf.UploadService)

		db.Init(&conf.DB)

		models.SetNotifier(notifier.New(conf.Unanswered))
		repository := models.NewConversationRepository(db.New())

		nats.Init(&conf.Nats, true)

		log.Info("Starting chat service on %s:%s", conf.Host, conf.Port)
		s := rpc.MakeServer(conf.Host + ":" + conf.Port)
		chat.RegisterChatServiceServer(s.Server, server.NewChatServer(
			repository,
		))
		s.StartServe()

		// interrupt
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		// wait for terminating
		<-interrupt
		log.Info("Service stopped")
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(cmdRun)
}
