package db

import (
	"fmt"
	"utils/log"

	"fetcher/conf"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"time"
)

var (
	DB *gorm.DB
)

func InitDB() {
	settings := conf.GetSettings()
	options := fmt.Sprintf(`
	user=%s
	password=%s
	host=%s
	port=%s
	dbname=%s
	sslmode=disable`,
		settings.DB.User,
		settings.DB.Password,
		settings.DB.Host,
		settings.DB.Port,
		settings.DB.Name)

	var (
		db  *gorm.DB
		err error
	)

	for {
		db, err = gorm.Open("postgres", options)
		if err == nil {
			err = db.DB().Ping()
		}

		if err == nil {
			break
		}

		log.Error(err)
		<-time.After(time.Second)
	}

	db.LogMode(settings.DB.Debug)
	DB = db
}
