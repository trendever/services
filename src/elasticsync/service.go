package main

import (
	"elasticsync/config"
	"elasticsync/models"
	"elasticsync/sync"
	"github.com/spf13/cobra"
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

			log.PanicLogger(sync.Loop)
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

	// @TODO command to compare sync table in db with actual index
	// may be helpful in case of problems with es cluster

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
