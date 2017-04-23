package main

import (
	"golang.org/x/net/context"
	"proto/telegram"
	"utils/nats"
	"utils/rpc"
)

// Init initializes telegram RPC service
func InitViews(t *Telegram) {

	grpcServer := rpc.Serve(GetSettings().RPC)

	telegram.RegisterTelegramServiceServer(grpcServer, telebotServer{
		Telegram: t,
	})

	nats.StanSubscribe(&nats.StanSubscription{})
}

type telebotServer struct {
	Telegram *Telegram
}

func (t telebotServer) NotifyMessage(ctx context.Context, req *telegram.NotifyMessageRequest) (*telegram.NotifyMessageResult, error) {
	go t.Telegram.Notify(req.Channel, req.Message)

	return &telegram.NotifyMessageResult{}, nil
}
