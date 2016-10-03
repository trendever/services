package models

import (
	"fetcher/conf"
	"utils/db"
	"utils/log"
)

var modelsList = []interface{}{
	&Activity{},
	&ThreadInfo{},
}

// AutoMigrate used models
func AutoMigrate(drop bool) error {
	// initialize database
	db.Init(&conf.GetSettings().DB)

	if drop {
		err := db.New().DropTableIfExists(modelsList...).Error
		if err != nil {
			return err
		}

		log.Warn("Drop Tables: success.")
	}

	err := db.New().AutoMigrate(modelsList...).Error
	if err != nil {
		return err
	}

	log.Info("Migration: success.")

	return nil
}
