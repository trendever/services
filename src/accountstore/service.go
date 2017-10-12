package main

import (
	"common/config"
	"common/db"
	"common/log"
	"instagram"
	"proto/core"
	"utils/cli"
	"utils/nats"
)

var service *svc

type svc struct {
	ig         InstagramAccess
	shopClient core.ShopServiceClient
}

func main() {
	service = &svc{ // use real db adapter anywhere but not in tests
		ig: &InstagramAccessImpl{},
	}
	cli.Main(service)
}

func (s *svc) Load() {
	config.LoadStruct("accountstore", &settings)
	log.Init(settings.Debug, "accountstore", settings.SentryDSN)
	db.Init(settings.DB)
	go notifier()
	nats.Init(&settings.Nats, true)
}

func (s *svc) Start() {
	s.Load()
	s.StartServer()
}

func (s *svc) Cleanup() {
}

func (s *svc) Migrate(drop bool) {

	s.Load()

	var models = []interface{}{
		&Account{},
	}

	if drop {
		db.New().DropTable(models...)
	}
	log.Error(db.New().AutoMigrate(models...).Error)
	fixIDs()
	idShouldBePrimary()
	log.Error(db.New().Exec("UPDATE accounts SET instagram_username = LOWER(instagram_username)").Error)
}

func fixIDs() {
	var broken []Account
	err := db.New().Where("instagram_id = 0").Where("valid").Find(&broken).Error
	if err != nil {
		log.Errorf("failed to load accounts with zero id: %v", err)
		return
	}
	for _, acc := range broken {
		ig, _ := instagram.Restore(acc.Cookie, "", false, false)
		if ig.UserID != 0 {
			acc.InstagramID = ig.UserID
			err := db.New().Save(&acc).Error
			if err != nil {
				log.Errorf("failed to save fixed acc: %v", err)
			}
		}
	}
}

func idShouldBePrimary() {
	tx := db.NewTransaction()
	log.Error(tx.Where(`instagram_id IN (SELECT instagram_id FROM accounts GROUP BY instagram_id HAVING COUNT(1) > 1) AND NOT valid`).Delete(Account{}).Error)
	log.Error(tx.Exec("ALTER TABLE accounts DROP CONSTRAINT accounts_pkey").Error)
	log.Error(tx.Exec("ALTER TABLE accounts ADD PRIMARY KEY (instagram_id)").Error)
	log.Error(tx.Model(&Account{}).AddIndex("idx_accounts_instagram_username", "instagram_username").Error)
	log.Error(tx.Commit().Error)
}
