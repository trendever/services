package models

import (
	"utils/db"
	"utils/log"
)

var dbModels = []interface{}{
	&PushNotify{},
	&PushReceiver{},
}

func Migrate(drop bool) {
	db := db.New()
	if drop {
		log.Warn("Droping tables")
		db.DropTableIfExists(dbModels...)
	}

	if err := db.AutoMigrate(dbModels...).Error; err != nil {
		log.Fatal(err)
	}
	db.Table("push_notifies_receivers").AddForeignKey("push_notify_id", "push_notifies(id)", "CASCADE", "RESTRICT")
	db.Table("push_notifies_receivers").AddForeignKey("push_receiver_id", "push_receivers(id)", "CASCADE", "RESTRICT")
}
