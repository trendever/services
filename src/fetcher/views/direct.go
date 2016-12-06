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

func (s fetcherServer) CreateThread(ctx context.Context, in *bot.CreateThreadRequest) (*bot.CreateThreadReply, error) {
	// @TODO timeout?
	tid, err := fetcher.CreateThread(in.Inviter, in.Participant, in.Caption, in.InitMessage)
	if err != nil {
		return &bot.CreateThreadReply{Error: err.Error()}, nil
	}
	return &bot.CreateThreadReply{ThreadId: tid}, nil
}
