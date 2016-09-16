package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/trendcoin"
	"syscall"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/rpc"
)

const ServiceName = "trendcoin"

var settings struct {
	Debug     bool
	RPC       string
	DB        db.Settings
	SentryDSN string
}

func init() {
	err := config.LoadStruct(ServiceName, &settings)
	log.Init(settings.Debug, ServiceName, settings.SentryDSN)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load config: %v", err))
	}
}

func main() {
	var cmd = cobra.Command{
		Use:   "service",
		Short: "trendcoin service",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting service...")

			rpc := rpc.Serve(settings.RPC)
			db.Init(&settings.DB)
			server := NewTrendcoinServer()

			log.Info("Registering server...")
			trendcoin.RegisterTrendcoinServiceServer(rpc, server)

			interrupt := make(chan os.Signal)
			signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

			<-interrupt
			rpc.Stop()
			server.Stop()
		},
	})

	var drop bool
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Runs database migration",
		Run: func(cmd *cobra.Command, args []string) {
			log.Warn("Starting database migration...")
			db.Init(&settings.DB)

			Migrate(drop)

			log.Info("Migration done")
		},
	}
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	cmd.AddCommand(migrateCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Runs tests",
		Run: func(cmd *cobra.Command, args []string) {
			// @TODO
		},
	})

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
