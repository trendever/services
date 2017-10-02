package cmd

import (
	"chat/config"
	"chat/models"
	"common/db"
	"github.com/spf13/cobra"
	"log"
)

var dbModels = []interface{}{
	models.Conversation{},
	models.Member{},
	models.Message{},
	models.MessagePart{},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs database migration",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		log.Println("Start database migration")
		db.Init(&config.Get().DB)
		db := db.New()

		if drop {
			log.Println("Drop tables")
			db.DropTableIfExists(dbModels...)
		}

		if err := db.AutoMigrate(dbModels...).Error; err != nil {
			log.Fatalf("Error during migration: %v", err)
		}
		if err := models.Migrate(); err != nil {
			log.Fatalf("Error during migration: %v", err)
		}
		log.Println("Migration done")

	},
}

var drop bool

func init() {
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	RootCmd.AddCommand(migrateCmd)
}
