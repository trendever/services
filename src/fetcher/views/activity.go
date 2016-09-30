package views

import (
	"fetcher/api"
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

func encodeActivity(act *models.Activity) *bot.Activity {
	return &bot.Activity{
		Id:                int64(act.ID),
		Pk:                act.Pk,
		MediaId:           act.MediaID,
		MediaUrl:          act.MediaURL,
		UserId:            act.UserID,
		UserImageUrl:      act.UserImageURL,
		UserName:          act.UserName,
		MentionedUsername: act.MentionedUsername,
		Type:              act.Type,
		Comment:           act.Comment,
		CreatedAt:         act.CreatedAt.Unix(),
	}
}

func encodeActivities(activities []models.Activity) []*bot.Activity {

	out := make([]*bot.Activity, len(activities))

	for i := range activities {
		out[i] = encodeActivity(&activities[i])
	}

	return out
}

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
		Result: encodeActivities(result),
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

	// do notify
	// @TODO

	// set notified
	err := db.New().
		Model(&models.ThreadInfo{}).
		Where("thread_id = ?", threadID).
		Update("notified", true).
		Error

	return &bot.SendDirectReply{}, nil
}
