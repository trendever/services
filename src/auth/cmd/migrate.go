package cmd

import (
	"auth/config"
	"auth/models"
	"github.com/spf13/cobra"
	"utils/db"
	"utils/log"
)

var dbModels = []interface{}{
	models.UserPassword{},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs database migration",
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Start database migration")
		db.Init(&config.Get().DB)
		if drop {
			log.Warn("Drop tables")
			err := db.New().DropTableIfExists(dbModels...).Error
			if err != nil {
				log.Fatal(err)
			}
		}
		err := db.New().AutoMigrate(dbModels...).Error
		if err != nil {
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
