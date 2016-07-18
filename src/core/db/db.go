package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"utils/log"
	"core/conf"
	"time"

	// connect to postgres
	_ "github.com/lib/pq"
)

var db *gorm.DB

// New returns new ready-to-use query object
func New() *gorm.DB {
	return db
}

// Init connection to database
func Init() {

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
	newdb.DB().SetMaxOpenConns(50)
	newdb.DB().SetMaxIdleConns(10)
	db = newdb
}
