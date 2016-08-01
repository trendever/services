package cmd

import (
	"os"
	"os/signal"
	"payments/config"
	"payments/db"
	"payments/views"
	"syscall"
	"utils/log"

	"github.com/spf13/cobra"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting payment service on %q", config.Get().RPC)

		config.Init()
		db.Init()
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
