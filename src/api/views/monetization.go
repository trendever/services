package views

import (
	"api/api"
	"api/soso"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"net/http"
	"proto/core"
	"proto/payment"
	"time"
	"utils/rpc"
)

var monetizationServiceClient = core.NewMonetizationServiceClient(api.CoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"get_plan", "monetization", GetMonetizationPlan},
		soso.Route{"plans_list", "monetization", GetMonetizationPlansList},
		soso.Route{"coins_offers", "monetization", GetCoinsOffers},
		soso.Route{"buy_coins", "monetization", BuyCoins},
		soso.Route{"subscribe", "monetization", SubscribeToPlan},
		soso.Route{"set_autorefill", "monetization", SetAutorefill},
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

	plan, err := getMonetizationPlan(uint64(id))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(plan)
}

func getMonetizationPlan(id uint64) (*core.MonezationPlan, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.GetPlan(ctx, &core.GetPlanRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, errors.New(res.Error)
	}
	return res.Plan, nil
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

	res, err := getCoinsOffers(currency, 0)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(res)
}

func SetAutorefill(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	offerID, _ := c.RequestMap["offer_id"].(float64)
	if offerID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid offer id"))
		return
	}

	balance, err := coinsBalance(c.Token.UID)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	if balance <= 0 {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("you should have positive balance to do this"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.SetAutorefill(ctx, &core.SetAutorefillRequest{
		UserId:  c.Token.UID,
		OfferId: uint64(offerID),
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.Error != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.Error))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

func BuyCoins(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	offerID, _ := c.RequestMap["offer_id"].(float64)
	if offerID <= 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("invalid offer id"))
		return
	}

	gateway, _ := c.RequestMap["gateway"].(string)
	if gateway == "" {
		gateway = "payture"
	}

	offersResp, err := getCoinsOffers("", uint64(offerID))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if len(offersResp.Offers) == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("unknown offer id"))
		return
	}
	offer := offersResp.Offers[0]

	currency, ok := payment.Currency_value[offer.Currency]
	if !ok {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New("unsupported currency"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	payResp, err := paymentServiceClient.CreateOrder(ctx, &payment.CreateOrderRequest{
		Data: &payment.OrderData{
			Amount:      uint64(offer.Amount),
			Currency:    payment.Currency(currency),
			Gateway:     gateway,
			ServiceName: "coins_refill",
			ServiceData: fmt.Sprintf(`{"user_id": %v, "amount": %v}`, c.Token.UID, offer.Amount),
			Comment:     fmt.Sprintf("%v trendcoins", offer.Amount),
		},
		Info: &payment.UserInfo{
			UserId: c.Token.UID,
		},
	})

	if err != nil { // RPC errors
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if payResp.Error > 0 { // service errors
		c.Response.ResponseMap = map[string]interface{}{
			"ErrorCode":    payResp.Error,
			"ErrorMessage": payResp.ErrorMessage,
		}
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(payResp.ErrorMessage))
		return
	}
	c.SuccessResponse(map[string]interface{}{
		"order_id": payResp.Id,
	})
}

func getCoinsOffers(currency string, id uint64) (*core.GetCoinsOffersReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := monetizationServiceClient.GetCoinsOffers(ctx, &core.GetCoinsOffersRequest{
		Currency: currency,
		OfferId:  id,
	})
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, errors.New(res.Error)
	}
	return res, nil
}
