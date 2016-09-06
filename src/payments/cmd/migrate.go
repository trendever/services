package cmd

import (
	"github.com/spf13/cobra"
	"payments/config"
	"payments/models"
	"utils/db"
	"utils/log"
)

var dbModels = []interface{}{
	&models.Payment{},
	&models.Session{},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs database migration",
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Starting database migration for payment service")
		config.Init()
		db.Init(&config.Get().DB)
		db := db.New()

		if drop {
			log.Warn("Droping tables")
			db.DropTableIfExists(dbModels...)
		}

		if err := db.AutoMigrate(dbModels...).Error; err != nil {
			log.Fatal(err)
		}
		if err := models.Migrate(db); err != nil {
			log.Fatal(err)
		}
		log.Info("Migration done")

	},
}

var drop bool

func init() {
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	RootCmd.AddCommand(migrateCmd)
}
