package views

import (
	"core/api"
	"core/models"
	"core/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"proto/trendcoin"
	"time"
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

func (s *monetizationServer) Subscribe(_ context.Context, in *core.SubscribeRequest) (ret *core.SubscribeReply, _ error) {
	ret = &core.SubscribeReply{}

	var shop models.Shop
	res := db.New().First(&shop, "id = ?", in.ShopId)
	if res.RecordNotFound() {
		ret.Error = "shop not found"
		return
	}
	if res.Error != nil {
		log.Errorf("failed to load shop: %v", res.Error)
		ret.Error = "db error"
		return
	}

	// only supplier can set plan
	if uint64(shop.SupplierID) != in.UserId {
		ret.Error = "unsuitable user"
	}

	var plan models.MonetizationPlan
	res = db.New().First(&plan, "id = ?", in.PlanId)
	if res.RecordNotFound() {
		ret.Error = "plan not found"
		return
	}
	if res.Error != nil {
		log.Errorf("failed to load plan: %v", res.Error)
		ret.Error = "db error"
		return
	}

	// non-public plans can not be voted with this api
	if !plan.Public {
		ret.Error = "plan is not public"
		return
	}

	shop.PlanID = plan.ID
	if plan.SubscriptionPeriod != 0 {

	}
	shop.PlanExpiresAt = &time.Now().Add(time.Hour * 24 * time.Duration(plan.SubscriptionPeriod))
	shop.Suspended = false
	shop.AutoRenewal = in.AutoRenewal

	// for plans without subscription fee
	if plan.SubscriptionPrice == 0 {
		err := db.New().Save(&shop).Error
		if err != nil {
			log.Errorf("failed to save shop: %v", res.Error)
			ret.Error = "db error"
		} else {
			ret.Ok = true
		}
		return
	}

	err := utils.PerformTransactions(&trendcoin.TransactionData{
		Source:         uint64(in.UserId),
		Amount:         plan.SubscriptionPrice,
		AllowEmptySide: true,
		Reason:         "subscription fee",
	})
	switch err.Error() {
	case nil:

	case "Invalid source account", "Credit is not allowed for this transaction":
		log.Errorf("failed to perform transactions: %v", err)
		ret.Error = "insufficient funds"
		return

	default:
		log.Errorf("failed to perform transactions: %v", err)
		ret.Error = "temporarily unable to write-off coins"
		return
	}

	err = db.New().Save(&shop).Error
	// here comes troubles
	if err != nil {
		log.Errorf("failed to save shop after coins write off: %v!", res.Error)
		refundErr := utils.PerformTransactions(&trendcoin.TransactionData{
			Destination:    uint64(in.UserId),
			Amount:         plan.SubscriptionPrice,
			AllowEmptySide: true,
			Reason:         "failed subscription refund",
		})
		if refundErr != nil {
			// well... things going really bad
			// @CHECK what else can we do?
			// @TODO use rpc with guaranteed delivery?
			log.Errorf("failed to refund coins %v to %v: %v!", plan.SubscriptionPrice, in.UserId, refundErr)
			ret.Error = "db error after coins write-off"
		} else {
			ret.Error = "db error"
		}
		return
	}

	ret.Ok = true
	return
}
