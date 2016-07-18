package db

import (
	"github.com/jinzhu/gorm"
	"utils/log"
	"time"
)

//GetPG returns database instance
func GetPG(config string) (*gorm.DB, error) {
	for {
		var db *gorm.DB
		var err error
		if db, err = gorm.Open("postgres", config); err == nil {
			if err = db.DB().Ping(); err == nil {
				//db.LogMode(conf.Debug)
				return db, nil
			}
		}
		log.Warn("DB error: %v \n try reconnect after 1 second \n %v", err, config)
		<-time.After(time.Second)
	}
}
