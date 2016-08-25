package models

import (
	"utils/db"
	"utils/log"
)

var dbModels = []interface{}{
	&Message{},
	&Receiver{},
}

func Migrate(drop bool) {
	db := db.New()
	if drop {
		log.Warn("Droping tables")
		db.DropTableIfExists(dbModels)
	}

	if err := db.AutoMigrate(dbModels...).Error; err != nil {
		log.Fatal(err)
	}
}
