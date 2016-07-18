package cmd

import (
	"github.com/spf13/cobra"
	"utils/log"
	"os"
	"os/signal"
	"payments/config"
	"syscall"
)

var cmdRun = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting payment service on %q", config.Get().Listen)

		// interrupt
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

		// wait for terminating
		for {
			select {
			case <-interrupt:
				log.Info("Payment service stopped")
				os.Exit(0)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(cmdRun)
}
