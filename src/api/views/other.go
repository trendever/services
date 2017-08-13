package views

import (
	"api/api"
	"api/conf"
	"api/soso"
	"errors"
	"net/http"
	"proto/checker"
	"proto/sms"
	"utils/phone"
	"utils/rpc"
)

var smsServiceClient = sms.NewSmsServiceClient(api.SMSConn)
var checkerServiceClient = checker.NewCheckerServiceClient(api.CheckerConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"common", "market_sms", SendMarketSMS},
		soso.Route{"instagram", "get_profile", GetInstagramProfile},
	)
}

func SendMarketSMS(c *soso.Context) {
	req := c.RequestMap
	phoneNumber, ok := req["phone"].(string)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Empty phone"))
		return
	}
	country, _ := req["country"].(string)
	phone, err := phone.CheckNumber(phoneNumber, country)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Invalid phone"))
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := smsServiceClient.SendSMS(ctx, &sms.SendSMSRequest{
		Phone: phone,
		Msg:   conf.GetSettings().MarketSMS,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if res.SmsError != "" {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, errors.New(res.SmsError))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "success",
	})
}

func GetInstagramProfile(c *soso.Context) {
	id, _ := c.RequestMap["id"].(float64)
	name, _ := c.RequestMap["name"].(string)
	if id < 1 && name == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("empty conditions"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := checkerServiceClient.GetProfile(ctx, &checker.GetProfileRequest{
		Id:   uint64(id),
		Name: name,
	})
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}

	c.SuccessResponse(res)
}
