package main

import (
	"golang.org/x/net/context"
	"proto/bot"
	"utils/rpc"
)

// Init initializes telegram RPC service
func InitViews(telegram *Telegram) {

	grpcServer := rpc.Serve(GetSettings().RPC)

	bot.RegisterTelegramServiceServer(grpcServer, telebotServer{
		Telegram: telegram,
	})
}

type telebotServer struct {
	Telegram *Telegram
}

func (t telebotServer) NotifyMessage(ctx context.Context, req *bot.NotifyMessageRequest) (*bot.NotifyMessageResult, error) {
	go t.Telegram.Notify(req.Channel, req.Message)

	return &bot.NotifyMessageResult{}, nil
}
