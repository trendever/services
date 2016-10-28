package cmd

import (
	"os"
	"os/signal"
	//"payments/api"
	"payments/config"
	"payments/views"
	"syscall"
	"utils/db"
	"utils/log"

	"github.com/spf13/cobra"
	"utils/nats"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting payment service on %q", config.Get().RPC)

		config.Init()

		//	api.Init()
		db.Init(&config.Get().DB)
		nats.Init(&config.Get().Nats, true)
		views.Init()

		// interrupt
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		<-interrupt
		log.Info("Payment service stopped")
		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(cmdRun)
}
