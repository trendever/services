package main

import (
	"golang.org/x/net/context"
	"proto/telegram"
	"time"
	"utils/nats"
	"utils/rpc"
)

// Init initializes telegram RPC service
func InitViews(t *Telegram) {

	grpcServer := rpc.MakeServer(GetSettings().RPC)
	srv := telebotServer{
		Telegram: t,
	}

	telegram.RegisterTelegramServiceServer(grpcServer.Server, srv)
	grpcServer.StartServe()

	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "telegram.notify",
		Group:          "telegram",
		DurableName:    "telegram",
		AckTimeout:     time.Second * 30,
		DecodedHandler: srv.StanMessage,
	})
	nats.Init(&settings.Nats, true)
}

type telebotServer struct {
	Telegram *Telegram
}

func (t telebotServer) NotifyMessage(ctx context.Context, req *telegram.NotifyMessageRequest) (*telegram.NotifyMessageResult, error) {
	err, _ := t.Telegram.Notify(req)
	if err != nil {
		return &telegram.NotifyMessageResult{Error: err.Error()}, nil
	} else {
		return &telegram.NotifyMessageResult{}, nil
	}
}

func (t telebotServer) StanMessage(req *telegram.NotifyMessageRequest) bool {
	_, retry := t.Telegram.Notify(req)
	return !retry
}
