package project

import (
	"fmt"
	"net/http"

	. "api/conf"
	"api/views"

	"utils/metrics"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"api/cache"
	"api/soso"
	"api/subscriber"
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
	settings := GetSettings()
	metrics.Init(settings.Metrics.Addr, settings.Metrics.User, settings.Metrics.Password, settings.Metrics.DBName)
	cache.Init()
	SosoObj.HandleList(views.SocketRoutes)
	subscriber.Init()
	return http.ListenAndServe(
		fmt.Sprintf(":%s", settings.ChannelPort),
		GetMainHandler(),
	)
}
