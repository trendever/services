package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"utils/log"
	"auth/db"
	"auth/models"
)

var dbModels = []interface{}{
	models.UserPassword{},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs database migration",
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Start database migration")
		db, err := db.GetPG(viper.GetString("db.config"))
		if err != nil {
			log.Fatal(err)
		}
		if drop {
			log.Warn("Drop tables")
			db.DropTableIfExists(dbModels...)
		}
		db.AutoMigrate(dbModels...)
		if db.Error != nil {
			log.Fatal(db.Error)
		}

		log.Info("Migration done")

	},
}

var drop bool

func init() {
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	RootCmd.AddCommand(migrateCmd)
}
