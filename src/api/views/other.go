package views

import (
	"api/api"
	"api/conf"
	"api/soso"
	"errors"
	"net/http"
	"proto/sms"
	"utils/rpc"
	"utils/phone"
)

var smsServiceClient = sms.NewSmsServiceClient(api.SMSConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"market_sms", "common", SendMarketSMS},
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
