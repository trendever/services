package cmd

import (
	"common/db"
	"common/log"
	"payments/config"
	"payments/models"
)

var dbModels = []interface{}{
	&models.Payment{},
	&models.Session{},
}

func (s *Service) Migrate(drop bool) {
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

}
