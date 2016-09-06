package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"sms/conf"
	"sms/models"
	"utils/db"
)

var (
	modelsList = []interface{}{
		&models.SmsDB{},
	}
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		db.Init(&conf.GetSettings().DB)
		if drop {
			err := db.New().DropTableIfExists(modelsList...).Error
			if err != nil {
				log.Fatalf("Can't drop tables: %v", err)
			}

			log.Println("Drop Tables: success.")
		}

		if err := db.New().AutoMigrate(modelsList...).Error; err != nil {
			log.Fatalf("Can't migreate tables: %v", err)
		}

		log.Println("Migration: success.")
	},
}

var (
	drop bool
)

func init() {
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drop tables before database migration")
	RootCmd.AddCommand(migrateCmd)
}
