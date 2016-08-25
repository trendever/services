package main

import (
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/push"
	"push/config"
	"push/exteral"
	"push/models"
	"push/server"
	"syscall"
	"utils/db"
	"utils/log"
	"utils/rpc"
)

func main() {
	var cmd = cobra.Command{
		Use:   "service",
		Short: "push service",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting service...")

			config.Init()
			exteral.Init()
			db.Init(&config.Get().DB)
			rpc := rpc.Serve(config.Get().RPC)
			server := server.NewPushServer()

			log.Info("Registering server...")
			push.RegisterPushServiceServer(
				rpc,
				server,
			)

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
			log.Warn("Starting database migration for elasticsync service")
			config.Init()
			c := config.Get()
			db.Init(&c.DB)

			models.Migrate(drop)

			log.Info("Migration done")
		},
	}
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	cmd.AddCommand(migrateCmd)

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
