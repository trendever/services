package views

import (
	"core/api"
	"core/models"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"utils/db"
	"utils/log"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		core.RegisterMonetizationServiceServer(s, &monetizationServer{})
	})
}

type monetizationServer struct{}

func (s *monetizationServer) GetPlan(_ context.Context, in *core.GetPlanRequest) (*core.GetPlanReply, error) {
	var plan models.MonetizationPlan
	err := db.New().First(&plan, "id = ?", in.Id).Error
	if err != nil {
		log.Errorf("failed to load monetization plan %v: %v", in.Id, err)
		return &core.GetPlanReply{Error: err.Error()}, nil
	}
	return &core.GetPlanReply{Plan: plan.Encode()}, nil
}

func (s *monetizationServer) GetPlansList(_ context.Context, in *core.GetPlansListRequest) (*core.GetPlansListReply, error) {
	var plans []models.MonetizationPlan
	scope := db.New()
	if in.Currency != "" {
		scope = scope.Where("primary_currency = ?", in.Currency)
	}
	err := scope.Find(&plans, "public").Error
	if err != nil {
		log.Errorf("failed to load monetization plans: %v", err)
		return &core.GetPlansListReply{Error: err.Error()}, nil
	}
	if len(plans) == 0 {
		return &core.GetPlansListReply{Error: "public plans not found"}, nil
	}
	ret := &core.GetPlansListReply{}
	for _, plan := range plans {
		ret.Plans = append(ret.Plans, plan.Encode())
	}
	return ret, nil
}

func (s *monetizationServer) GetCoinsOffers(_ context.Context, in *core.GetCoinsOffersRequest) (*core.GetCoinsOffersReply, error) {
	var offers []models.CoinsOffer
	scope := db.New()
	if in.Currency != "" {
		scope = scope.Where("currency = ?", in.Currency)
	}
	err := scope.Find(&offers).Error
	if err != nil {
		log.Errorf("failed to load coins offers: %v", err)
		return &core.GetCoinsOffersReply{Error: err.Error()}, nil
	}
	ret := &core.GetCoinsOffersReply{}
	for _, offer := range offers {
		ret.Offers = append(ret.Offers, offer.Encode())
	}
	return ret, nil
}
