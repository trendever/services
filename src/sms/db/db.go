package db

import (
	"fmt"
	"utils/log"

	"sms/conf"

	"github.com/jinzhu/gorm"
	"time"
)

var (
	//DB instance
	DB *gorm.DB
)

//InitDB initializes database instance
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
		newdb *gorm.DB
		err   error
	)
	for { // wait until connection to database
		log.Debug(options)
		newdb, err = gorm.Open("postgres", options)
		if err == nil {
			err = newdb.DB().Ping()
		}

		if err == nil {
			break
		}

		log.Warn("Database connection error (%v)! Retrying in 1 second\n", err.Error())
		<-time.After(time.Second)
	}

	newdb.LogMode(settings.DB.Debug)
	DB = newdb
}
