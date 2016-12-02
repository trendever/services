package views

import (
	"fetcher/fetcher"
	"golang.org/x/net/context"
	"proto/bot"
)

// SendDirect sends message to the chat (if not sent earlier)
func (s fetcherServer) SendDirect(ctx context.Context, in *bot.SendDirectRequest) (*bot.SendDirectReply, error) {
	go fetcher.SendDirect(in.SenderId, in.RecieverId, in.ThreadId, in.Text)
	return &bot.SendDirectReply{}, nil
}

func (s fetcherServer) CreateThread(context.Context, *bot.CreateThreadRequest) (*bot.CreateThreadReply, error) {
	return &bot.CreateThreadReply{Error: "not impemented"}, nil
}
