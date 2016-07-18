package views

import (
	"telebot/conf"
	"telebot/telegram"

	"proto/bot"
	"utils/log"
	"utils/rpc"
	"golang.org/x/net/context"
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
	go func() {
		err := t.Telegram.Notify(req.Channel, req.Message)
		if err != nil {
			log.Error(err)
		}
	}()

	return &bot.NotifyMessageResult{}, nil
}
