package views

import (
	"fetcher/api"
	"fetcher/models"
	"golang.org/x/net/context"
	"proto/bot"
	"utils/db"
	"utils/log"
)

func Init() {
	bot.RegisterFetcherServiceServer(api.GrpcServer, fetcherServer{})
}

type fetcherServer struct{}

func encodeActivity(act *models.Activity) *bot.Activity {
	return &bot.Activity{
		Id:                int64(act.ID),
		Pk:                act.Pk,
		MediaId:           act.MediaID,
		MediaUrl:          act.MediaUrl,
		UserId:            act.UserID,
		UserImageUrl:      act.UserImageUrl,
		UserName:          act.UserName,
		MentionedUsername: act.MentionedUsername,
		Type:              act.Type,
		Comment:           act.Comment,
		CreatedAt:         act.CreatedAt.Unix(),
	}
}

func encodeActivities(activities []models.Activity) []*bot.Activity {

	out := make([]*bot.Activity, len(activities), len(activities))

	for i := range activities {
		out[i] = encodeActivity(&activities[i])
	}

	return out
}

// Returns activity (oldest first) for User since Timestamp with type Type
func (s fetcherServer) RetrieveActivities(ctx context.Context, in *bot.RetrieveActivitiesRequest) (*bot.RetrieveActivitiesResult, error) {

	result := []models.Activity{}

	req := db.New().
		Order("updated_at asc").
		Limit(int(in.Limit))

	var search_activity = models.Activity{
		MentionedUsername: in.MentionName,
		Type:              in.Type,
	}

	if in.AfterId > 0 {
		req = req.Where("id > ?", in.AfterId)
	}

	if err := req.Where(&search_activity).Find(&result).Error; err != nil {
		log.Error(err)
		return nil, err
	}

	return &bot.RetrieveActivitiesResult{
		Result: encodeActivities(result),
	}, nil
}
