package db

import (
	"github.com/jinzhu/gorm"
	"payments/config"
	"time"
	"utils/log"
)

var connection *gorm.DB

// New database conn
func New() *gorm.DB {
	return connection
}

// Init initializes db connection
func Init() {
	conf := config.Get()
	for {
		var db *gorm.DB
		var err error
		if db, err = gorm.Open("postgres", conf.DB.Conf); err == nil {
			if err = db.DB().Ping(); err == nil {
				db.LogMode(conf.DB.Debug)
				connection = db
				return
			}
		}
		log.Warn("DB error: %v \n try reconnect after 1 second \n %v", err, conf.DB.Conf)
		<-time.After(time.Second)
	}
}
