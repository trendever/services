package views

import (
	"fetcher/api"
	"fetcher/fetcher"
	"fetcher/models"

	"golang.org/x/net/context"

	"proto/bot"
	"utils/db"
	"utils/log"
)

// Init binds server
func Init() {
	bot.RegisterFetcherServiceServer(api.GrpcServer, fetcherServer{})
}

type fetcherServer struct{}

// Returns activity (oldest first) for User since Timestamp with type Type
func (s fetcherServer) RetrieveActivities(ctx context.Context, in *bot.RetrieveActivitiesRequest) (*bot.RetrieveActivitiesReply, error) {

	result := []models.Activity{}

	req := db.New().
		Order("updated_at asc").
		Limit(int(in.Limit))

	var searchActivity = models.Activity{
		MentionedUsername: in.MentionName,
		Type:              in.Type,
	}

	if in.AfterId > 0 {
		req = req.Where("id > ?", in.AfterId)
	}

	if err := req.Where(&searchActivity).Find(&result).Error; err != nil {
		log.Error(err)
		return nil, err
	}

	return &bot.RetrieveActivitiesReply{
		Result: models.EncodeActivities(result),
	}, nil
}

// SendDirect sends message to the chat (if not sent earlier)
func (s fetcherServer) SendDirect(ctx context.Context, in *bot.SendDirectRequest) (*bot.SendDirectReply, error) {

	// find thread info
	var info models.ThreadInfo
	err := db.New().Where("thread_id = ?", in.ThreadId).Find(&info).Error
	if err != nil {
		return nil, err
	}

	if info.Notified { // all ok; do nothing
		return &bot.SendDirectReply{}, nil
	}

	// find related activity to get bot username
	var act models.Activity
	err = db.New().Where("thread_id = ?", in.ThreadId).Find(&act).Error
	if err != nil {
		return nil, err
	}

	worker, err := fetcher.GetWorker(act.MentionedUsername)
	if err != nil {
		return nil, err
	}

	err = worker.SendDirectMsg(in.ThreadId, in.Text)
	if err != nil {
		log.Debug("Could not send shiet: %v", err)
		return nil, err
	}

	// set notified
	err = db.New().
		Model(&models.ThreadInfo{}).
		Where("thread_id = ?", info.ThreadID).
		Update("notified", true).
		Error

	return &bot.SendDirectReply{}, nil
}
