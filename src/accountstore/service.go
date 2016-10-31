package main

import (
	"instagram"
	"utils/cli"
	"utils/config"
	"utils/db"
	"utils/log"
)

var service *svc

type svc struct {
	repo AccountRepo
	ig   InstagramAccess
}

func main() {
	service = &svc{ // use real db adapter anywhere but not in tests
		repo: &AccountRepoImpl{},
		ig:   &InstagramAccessImpl{},
	}
	cli.Main(service)
}

func (s *svc) Load() {
	config.LoadStruct("accountstore", &settings)
	log.Init(settings.Debug, "accountstore", settings.SentryDSN)
	db.Init(settings.DB)
	instagram.DoResponseLogging = settings.InstagramDebug
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

	if !drop {
		db.New().AutoMigrate(models...)
	} else {
		db.New().DropTable(models...)
	}
}
