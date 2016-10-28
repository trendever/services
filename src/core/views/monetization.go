package views

import (
	"core/api"
	"core/models"
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"proto/core"
	"time"
	"utils/coins"
	"utils/db"
	"utils/log"
)

func init() {
	api.AddOnStartCallback(func(s *grpc.Server) {
		server := &monetizationServer{}
		core.RegisterMonetizationServiceServer(s, server)
		go server.loop()
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

	now := time.Now()
	// @CHECK prolongation actuality may be repetitive. what about it?
	if now.Sub(shop.LastPlanUpdate) < time.Minute {
		ret.Error = "action may be repetitive"
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
	err := subscribe(&shop, &plan, in.AutoRenewal)
	if err != nil {
		ret.Error = err.Error()
	} else {
		ret.Ok = true
	}
	// @TODO notifications
	return
}

func subscribe(shop *models.Shop, plan *models.MonetizationPlan, autoRenewal bool) error {
	now := time.Now()
	updateMap := map[string]interface{}{
		"plan_id":          plan.ID,
		"suspended":        false,
		"auto_renewal":     autoRenewal,
		"last_plan_update": now,
	}
	if plan.SubscriptionPeriod != 0 {
		// prolongation
		if plan.ID == shop.PlanID && shop.PlanExpiresAt.After(now) {
			updateMap["plan_expires_at"] = shop.PlanExpiresAt.Add(models.PlansBaseDuration * time.Duration(plan.SubscriptionPeriod))
		} else {
			updateMap["plan_expires_at"] = now.Add(models.PlansBaseDuration * time.Duration(plan.SubscriptionPeriod))
		}
	} else {
		updateMap["plan_expires_at"] = time.Time{}
	}

	// for plans without subscription fee
	if plan.SubscriptionPrice == 0 {
		err := db.New().Model(shop).UpdateColumns(updateMap).Error
		if err != nil {
			log.Errorf("failed to save shop: %v", err)
			return errors.New("db error")
		}
		return nil
	}

	err := coins.CheckWriteOff(
		uint64(shop.SupplierID), plan.SubscriptionPrice, "subscription fee",
		func() error {
			return db.New().Model(&shop).UpdateColumn(updateMap).Error
		},
	)
	switch err {
	case coins.CallbackFailed:
		err = errors.New("db error")
	case coins.RefundError:
		err = errors.New("unrecoverable error, refund failed")
	}
	return err
}

func (s *monetizationServer) loop() {
	for now := range time.Tick(time.Minute) {
		log.Debug("checking subscriptions...")
		var shops []*models.Shop
		err := db.New().Preload("Plan").
			Where("plan_expires_at < ?", now).
			//ignore plans without expiration
			Where("plan_expires_at != ? AND plan_expires_at IS NOT NULL", time.Time{}).
			Where("NOT suspended").
			Find(&shops).Error
		if err != nil {
			log.Errorf("failed to load shops with expired subscriptions: %v", err)
			continue
		}
		for _, shop := range shops {
			// @TODO notifications
			if !shop.AutoRenewal {
				err := db.New().Model(shop).UpdateColumn("suspended", true).Error
				if err != nil {
					log.Errorf("failed to suspend shop: %v", err)
				}
				continue
			}
			err := subscribe(shop, &shop.Plan, true)
			switch {
			case err == nil:
			case err.Error() == "insufficient funds":
				// @TODO autorefill coins
				log.Errorf("shop %v should be suspended due to not able to pay the subscription fee", shop.ID)
				err := db.New().Model(shop).UpdateColumn("suspended", true).Error
				if err != nil {
					log.Errorf("failed to suspend shop: %v", err)
				}
			default:
				log.Errorf("failed to renew subscription if shop %v: %v", shop.ID, err)
			}
		}
	}
}
