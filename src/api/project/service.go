package project

import (
	"api/auth"
	"api/cache"
	"api/conf"
	"api/views"
	"common/log"
	"common/metrics"
	"common/soso"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	"utils/elastic"
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
	soso.AddMiddleware(TokenMiddleware)
	metrics.Init(settings.Metrics.Addr, settings.Metrics.User, settings.Metrics.Password, settings.Metrics.DBName)
	cache.Init()
	SosoObj.HandleRoutes(views.SocketRoutes)
	nats.Init(&settings.Nats, true)
	elastic.Init(&settings.Elastic)
	http.Handle("/channel/", GetMainHandler())

	log.Info("Ready to listen")
	return http.ListenAndServe(
		fmt.Sprintf(":%s", settings.ChannelPort),
		nil,
	)
}

func TokenMiddleware(req *soso.Request, ctx *soso.Context, session soso.Session) error {
	if token, ok := req.TransMap["token"].(string); ok {
		tokenObj, err := auth.GetTokenData(token)
		if err != nil {
			return err
		}
		ctx.Token = &soso.Token{UID: tokenObj.UID, Exp: tokenObj.Exp}
	}

	return nil
}
