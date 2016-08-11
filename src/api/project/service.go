package project

import (
	"fmt"
	"net/http"

	. "api/conf"
	"api/views"

	"api/cache"
	"api/soso"
	"api/subscriber"
	"github.com/igm/sockjs-go/sockjs"
	"utils/metrics"
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
	http.Handle("/channel/", GetMainHandler())
	return http.ListenAndServe(
		fmt.Sprintf(":%s", settings.ChannelPort),
		nil,
	)
}
