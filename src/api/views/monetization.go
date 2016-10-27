package views

import (
	"api/api"
	"api/soso"
	"errors"
	"golang.org/x/net/context"
	"net/http"
	"proto/core"
	"time"
	"utils/rpc"
)

var monetizationServiceClient = core.NewMonetizationServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"get_plan", "monetization", GetMonetizationPlan},
		soso.Route{"plans_list", "monetization", GetMonetizationPlansList},
		soso.Route{"subscribe", "monetization", SubscribeToPlan},
	)
}

func SubscribeToPlan(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	planID, _ := c.RequestMap["plan_id"].(float64)
	if planID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid plan id"))
		return
	}
	shopID, _ := c.RequestMap["shop_id"].(float64)
	if shopID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid shop id"))
		return
	}
	autoRenewal, _ := c.RequestMap["auto_renewal"].(bool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*35)
	defer cancel()

	res, err := monetizationServiceClient.Subscribe(ctx, &core.SubscribeRequest{
		ShopId:      uint64(shopID),
		UserId:      c.Token.UID,
		PlanId:      uint64(planID),
		AutoRenewal: autoRenewal,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	// actuality may be not success, check its error field
	c.SuccessResponse(res)
}

func GetMonetizationPlan(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	id, _ := c.RequestMap["plan_id"].(float64)
	if id <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid plan id"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.GetPlan(ctx, &core.GetPlanRequest{Id: uint64(id)})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(res.Plan)
}

func GetMonetizationPlansList(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	currency, _ := c.RequestMap["currency"].(string)

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.GetPlansList(ctx, &core.GetPlansListRequest{Currency: currency})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(res)
}

func GetCoinsOffers(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	currency, _ := c.RequestMap["currency"].(string)

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.GetCoinsOffers(ctx, &core.GetCoinsOffersRequest{Currency: currency})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(res)
}
