package db

import (
	"github.com/jinzhu/gorm"
	"log"
	"chat/config"
	"time"
)

//GetPG returns db instance
func GetPG() *gorm.DB {
	conf := config.Get()
	for {
		var db *gorm.DB
		var err error
		if db, err = gorm.Open("postgres", conf.DB.Conf); err == nil {
			if err = db.DB().Ping(); err == nil {
				db.LogMode(conf.Debug)
				return db
			}
		}
		log.Printf("DB error: %v \n try reconnect after 1 second \n %v", err, conf.DB.Conf)
		<-time.After(time.Second)
	}
}
