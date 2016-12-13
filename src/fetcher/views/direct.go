package views

import (
	"fetcher/fetcher"
	"golang.org/x/net/context"
	"proto/bot"
	"sync"
	"time"
	"utils/log"
	"utils/nats"
)

const (
	SendDirectSubject = "direct.send"

	CreateThreadSubject      = "direct.create_thread"
	CreateThreadReplySubject = "direct.create_thread.reply"
)

var once sync.Once

func subscribe() {
	once.Do(func() {
		nats.StanSubscribe(
			&nats.StanSubscription{
				Subject:        SendDirectSubject,
				Group:          "fetcher",
				DurableName:    "fetcher",
				AckTimeout:     time.Second * 20,
				DecodedHandler: sendDirectNats,
			},
			&nats.StanSubscription{
				Subject:        CreateThreadSubject,
				Group:          "fetcher",
				DurableName:    "fetcher",
				AckTimeout:     time.Second * 20,
				DecodedHandler: createThread,
			},
		)
	})
}

func (s fetcherServer) SendDirect(ctx context.Context, in *bot.SendDirectRequest) (*bot.SendDirectReply, error) {
	go func() {
		_, err := fetcher.SendDirect(in.SenderId, in.RecieverId, in.ThreadId, in.Text)
		if err != nil {
			log.Errorf("failed to send message from %v: %v", in.SenderId, err)
		}
	}()
	return &bot.SendDirectReply{}, nil
}

func sendDirectNats(in *bot.SendDirectRequest) bool {
	mid, err := fetcher.SendDirect(in.SenderId, in.RecieverId, in.ThreadId, in.Text)
	reply := bot.DirectMessageNotify{ThreadId: in.ThreadId, ReplyKey: in.ReplyKey}
	switch err {
	case nil:
		reply.MessageId = mid
	case fetcher.AccountUnavailable:
		reply.Error = err.Error()
	default:
		log.Errorf("failed to send message from %v: %v", in.SenderId, err)
		// external trouble, try again later
		return false
	}
	// @TODO send it inside worker
	err = nats.StanPublish(fetcher.DirectMessageSubject, &reply)
	if err != nil {
		log.Errorf("failed to send reply via stan: %v", err)
		return false
	}
	return true
}

func createThread(in *bot.CreateThreadRequest) bool {
	tid, err := fetcher.CreateThread(in.Inviter, in.Participant, in.Caption, in.InitMessage)
	reply := bot.CreateThreadReply{ReplyKey: in.ReplyKey}
	switch err {
	case nil:
		reply.ThreadId = tid
	case fetcher.AccountUnavailable:
		reply.Error = err.Error()
	default:
		log.Errorf("failed to create thread from %v: %v", in.Inviter, err)
		// external trouble, try again later
		return false
	}
	err = nats.StanPublish(CreateThreadReplySubject, &reply)
	if err != nil {
		log.Errorf("failed to send reply via stan: %v", err)
		return false
	}
	return true
}
