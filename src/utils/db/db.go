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
				break
			}
		}
		log.Warn("DB error: %v \n try to reconnect after 5 seconds", err)
		time.Sleep(5 * time.Second)
	}
	connection.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("gorm:after_commit", afterCommitCallback)
	connection.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("gorm:after_commit", afterCommitCallback)
}

// afterCommitCallback will invoke `AfterCommit` method after commit
func afterCommitCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("AfterCommit")
	}
}
