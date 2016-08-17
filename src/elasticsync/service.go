package main

import (
	"elasticsync/config"
	"elasticsync/models"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils/db"
	"utils/elastic"
	"utils/log"
)

func main() {
	var cmd = cobra.Command{
		Use:   "service",
		Short: "elasticsync service",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting elasticsync service...")

			config.Init()
			c := config.Get()
			db.Init(&c.DB)
			elastic.Init(&c.Elastic)

			syncLoop()
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
			elastic.Init(&c.Elastic)

			models.Migrate(drop)

			log.Info("Migration done")
		},
	}
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables and elastic search index before migration")
	cmd.AddCommand(migrateCmd)

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}

func syncLoop() {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	conf := config.Get()
	for {

		select {
		case <-interrupt:
			log.Info("elasticsync service stopped")
			os.Exit(0)
		default:
			time.Sleep(time.Second * conf.Delay)
		}
	}
}
