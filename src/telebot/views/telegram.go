package views

import (
	"telebot/conf"
	"telebot/telegram"

	"golang.org/x/net/context"
	"proto/bot"
	"utils/rpc"
)

// Init initializes telegram RPC service
func Init(telegram *telegram.Telegram) {

	grpcServer := rpc.Serve(conf.GetSettings().RPC)

	bot.RegisterTelegramServiceServer(grpcServer, telebotServer{
		Telegram: telegram,
	})
}

type telebotServer struct {
	Telegram *telegram.Telegram
}

func (t telebotServer) NotifyMessage(ctx context.Context, req *bot.NotifyMessageRequest) (*bot.NotifyMessageResult, error) {
	go t.Telegram.Notify(req.Channel, req.Message)

	return &bot.NotifyMessageResult{}, nil
}
