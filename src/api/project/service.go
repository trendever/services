package project

import (
	"api/cache"
	"api/conf"
	"api/soso"
	"api/views"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	"utils/elastic"
	"utils/log"
	"utils/metrics"
	"utils/nats"

	// nats subscriptions
	_ "api/notifications"
)

// SosoObj is soso controller
var SosoObj = soso.Default()

// Receiver is sockjs to soso adapter
func Receiver(session sockjs.Session) {
	// Обработка входящих команд.
	SosoObj.RunReceiver(session)
}

// GetMainHandler is sockjs to http adapter
func GetMainHandler() http.Handler {
	return sockjs.NewHandler("/channel", sockjs.DefaultOptions, Receiver)
}

// Service main actions object
type Service struct{}

// Run main stuff
func (s *Service) Run() error {
	settings := conf.GetSettings()
	log.Info("Starting api service...")
	metrics.Init(settings.Metrics.Addr, settings.Metrics.User, settings.Metrics.Password, settings.Metrics.DBName)
	cache.Init()
	SosoObj.HandleList(views.SocketRoutes)
	nats.Init(&settings.Nats, false)
	elastic.Init(&settings.Elastic)
	http.Handle("/channel/", GetMainHandler())

	log.Info("Ready to listen")
	return http.ListenAndServe(
		fmt.Sprintf(":%s", settings.ChannelPort),
		nil,
	)
}
