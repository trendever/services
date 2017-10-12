package views

import (
	"common/db"
	"common/log"
	"errors"
	"fetcher/models"
	"golang.org/x/net/context"
	"proto/bot"
	"sync"
	"time"
	"utils/nats"
)

const (
	SendDirectSubject   = "direct.send"
	CreateThreadSubject = "direct.create_thread"
)

var once sync.Once

func subscribe() {
	once.Do(func() {
		nats.StanSubscribe(
			&nats.StanSubscription{
				Subject:        SendDirectSubject,
				Group:          "fetcher",
				DurableName:    "fetcher",
				AckTimeout:     time.Second * 10,
				DecodedHandler: saveRequest,
			},
			&nats.StanSubscription{
				Subject:        CreateThreadSubject,
				Group:          "fetcher",
				DurableName:    "fetcher",
				AckTimeout:     time.Second * 10,
				DecodedHandler: createThread,
			},
		)
	})
}

func (s fetcherServer) SendDirect(ctx context.Context, in *bot.SendDirectRequest) (*bot.SendDirectReply, error) {
	if saveRequest(in) {
		return &bot.SendDirectReply{}, nil
	} else {
		return nil, errors.New("failed to save request")
	}
}

func saveRequest(in *bot.SendDirectRequest) bool {
	var req = models.DirectRequest{
		Kind:     in.Type,
		UserID:   in.SenderId,
		ReplyKey: in.ReplyKey,
		ThreadID: in.ThreadId,
		Data:     in.Data,
	}
	if in.RecieverId != 0 {
		req.Participants = []uint64{in.RecieverId}
	}
	err := db.New().Save(&req).Error
	if err != nil {
		log.Errorf("failed to save request: %v", err)
		return false
	}
	return true
}

func createThread(in *bot.CreateThreadRequest) bool {
	err := db.New().Save(&models.DirectRequest{
		Kind:         bot.MessageType_CreateThread,
		UserID:       in.Inviter,
		ReplyKey:     in.ReplyKey,
		Participants: in.Participant,
		Caption:      in.Caption,
		Data:         in.InitMessage,
	}).Error
	if err != nil {
		log.Errorf("failed to save request: %v", err)
		return false
	}
	return true
}
