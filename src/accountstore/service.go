package main

import (
	"instagram"
	"proto/core"
	"proto/telegram"
	"utils/cli"
	"utils/config"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

var service *svc

type svc struct {
	ig         InstagramAccess
	shopClient core.ShopServiceClient
	teleClient telegram.TelegramServiceClient
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
	instagram.DoResponseLogging = settings.InstagramDebug
	global.telebotClient = telegram.NewTelegramServiceClient(rpc.Connect(settings.RPC.Telegram))
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
	db.New().AutoMigrate(models...)
	fixIDs()
}

func fixIDs() {
	var broken []Account
	err := db.New().Where("instagram_id = 0").Where("valid").Find(&broken).Error
	if err != nil {
		log.Errorf("failed to load accounts with zero id: %v", err)
		return
	}
	for _, acc := range broken {
		ig, _ := instagram.Restore(acc.Cookie, "", false)
		if ig.UserID != 0 {
			acc.InstagramID = ig.UserID
			err := db.New().Save(&acc).Error
			if err != nil {
				log.Errorf("failed to save fixed acc: %v", err)
			}
		}
	}
}
