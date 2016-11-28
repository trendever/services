package views

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"fetcher/models"
	"proto/bot"
	"utils/db"
	"utils/log"
)

// Init binds server
func Init(srv *grpc.Server) {
	bot.RegisterFetcherServiceServer(srv, fetcherServer{})
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
	}

	if len(in.Type) > 0 {
		req = req.Where("type in (?)", in.Type)
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
