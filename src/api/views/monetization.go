package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/core"
	"utils/rpc"
)

var monetizationServiceClient = core.NewMonetizationServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"get_plan", "monetization", GetMonetizationPlan},
		soso.Route{"plans_list", "monetization", GetMonetizationPlansList},
		soso.Route{"coins_offers", "monetization", GetCoinsOffers},
	)
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
