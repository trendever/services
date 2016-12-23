package views

import (
	"fetcher/models"
	"golang.org/x/net/context"
	"proto/bot"
	"sync"
	"time"
	"utils/db"
	"utils/log"
	"utils/nats"
)

const (
	SendDirectSubject   = "direct.send"
	CreateThreadSubject = "direct.create_thread"
)

var once sync.Once

var typeMap = map[bot.MessageType]models.RequestType{
	bot.MessageType_Text:       models.SendMessageRequest,
	bot.MessageType_MediaShare: models.ShareMediaRequest,
}

func subscribe() {
	once.Do(func() {
		nats.StanSubscribe(
			&nats.StanSubscription{
				Subject:        SendDirectSubject,
				Group:          "fetcher",
				DurableName:    "fetcher",
				AckTimeout:     time.Second * 10,
				DecodedHandler: addDirectRequest,
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
	addDirectRequest(in)
	return &bot.SendDirectReply{}, nil
}

func addDirectRequest(in *bot.SendDirectRequest) bool {
	var req = models.DirectRequest{
		Type:         typeMap[in.Type],
		UserID:       in.SenderId,
		ReplyKey:     in.ReplyKey,
		ThreadID:     in.ThreadId,
		Participants: []uint64{in.RecieverId},
		Data:         in.Data,
	}
	err := db.New().Save(&req).Error
	if err != nil {
		log.Errorf("failed to add direct request: %v", err)
		return false
	}
	return true
}

func createThread(in *bot.CreateThreadRequest) bool {
	err := db.New().Save(&models.DirectRequest{
		Type:         models.CreateThreadRequest,
		UserID:       in.Inviter,
		ReplyKey:     in.ReplyKey,
		Participants: in.Participant,
		Caption:      in.Caption,
		Data:         in.InitMessage,
	}).Error
	if err != nil {
		log.Errorf("failed to add direct request: %v", err)
		return false
	}
	return true
}
