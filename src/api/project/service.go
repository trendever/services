package project

import (
	"fmt"
	"net/http"

	"api/cache"
	"api/conf"
	"api/soso"
	"api/views"
	"github.com/igm/sockjs-go/sockjs"
	"utils/elastic"
	"utils/metrics"
	"utils/nats"
)

var SosoObj = soso.Default()

func Receiver(session sockjs.Session) {
	// Обработка входящих команд.
	SosoObj.RunReceiver(session)
}

func GetMainHandler() http.Handler {
	return sockjs.NewHandler("/channel", sockjs.DefaultOptions, Receiver)
}

type ProjectService struct{}

func (s *ProjectService) Run() error {
	settings := conf.GetSettings()
	metrics.Init(settings.Metrics.Addr, settings.Metrics.User, settings.Metrics.Password, settings.Metrics.DBName)
	cache.Init()
	SosoObj.HandleList(views.SocketRoutes)
	nats.Init(settings.NatsURL)
	elastic.Init(&settings.Elastic)
	http.Handle("/channel/", GetMainHandler())
	return http.ListenAndServe(
		fmt.Sprintf(":%s", settings.ChannelPort),
		nil,
	)
}
