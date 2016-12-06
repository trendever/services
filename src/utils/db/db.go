package db

import (
	"database/sql"
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

// similar to gorm.Model but with uint64 key
type Model struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// New database conn
func New() *gorm.DB {
	return connection.New()
}

func NilScan(rows *sql.Rows, args ...interface{}) (err error) {
	cp := make([]interface{}, len(args), len(args))

	type replaced struct {
		orig *string
		null sql.NullString
	}
	reps := make([]replaced, 0, len(args))
	for i, arg := range args {
		switch arg := arg.(type) {
		case *string:
			reps = append(reps, replaced{
				orig: arg,
			})
			cp[i] = &reps[len(reps)-1].null
		default:
			cp[i] = arg
		}
	}
	err = rows.Scan(cp...)
	if err != nil {
		return
	}

	for _, rep := range reps {
		*rep.orig = rep.null.String
	}
	return
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
		log.Warn("DB error: %v\n try to reconnect after 5 seconds", err)
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

var transactionCallback *gorm.Callback

// returns db with already started transaction
// and restricted version of callbacks to avoid nested transactions generated with default gorm callbacks
func NewTransaction() *gorm.DB {
	db := New()
	if transactionCallback == nil {
		transactionCallback = &gorm.Callback{}
		*transactionCallback = *gorm.DefaultCallback
		transactionCallback.Create().Remove("gorm:begin_transaction")
		transactionCallback.Create().Remove("gorm:commit_or_rollback_transaction")
		transactionCallback.Update().Remove("gorm:begin_transaction")
		transactionCallback.Update().Remove("gorm:commit_or_rollback_transaction")
		transactionCallback.Delete().Remove("gorm:begin_transaction")
		transactionCallback.Delete().Remove("gorm:commit_or_rollback_transaction")
		// there will be no commit actuality... but we still want to invoke our final callbacks
		transactionCallback.Create().Register("gorm:after_commit", afterCommitCallback)
		transactionCallback.Update().Register("gorm:after_commit", afterCommitCallback)
	}
	// looks like dirty hack
	*db.Callback() = *transactionCallback
	return db.Begin()
}
