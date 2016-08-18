package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
	"utils/log"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var connection *gorm.DB

type Settings struct {
	Host     string
	Port     uint
	User     string
	Password string
	Base     string
	Debug    bool
}

// New database conn
func New() *gorm.DB {
	return connection
}

// Init initializes db connection
func Init(s *Settings) {
	options := fmt.Sprintf(
		"host=%s port=%v user=%s password=%s dbname=%s sslmode=disable",
		s.Host,
		s.Port,
		s.User,
		s.Password,
		s.Base,
	)
	log.Info("DB options string: %v", options)
	for {
		var db *gorm.DB
		var err error
		if db, err = gorm.Open("postgres", options); err == nil {
			if err = db.DB().Ping(); err == nil {
				db.LogMode(s.Debug)
				connection = db
				return
			}
		}
		log.Warn("DB error: %v \n try reconnect after 1 second", err)
		time.Sleep(time.Second)
	}
}
