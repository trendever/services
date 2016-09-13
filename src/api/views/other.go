package views

import (
	"api/api"
	"api/conf"
	"api/soso"
	"errors"
	"net/http"
	"proto/sms"
	"utils/rpc"

	"github.com/ttacon/libphonenumber"
)

var smsServiceClient = sms.NewSmsServiceClient(api.SMSConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"market_sms", "common", SendMarketSMS},
	)
}

func SendMarketSMS(c *soso.Context) {
	req := c.RequestMap
	phone, ok := req["phone"].(string)
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Empty phone"))
		return
	}
	country, _ := req["country"].(string)
	phone, err := checkPhoneNumber(phone, country)
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

func checkPhoneNumber(phoneNumber, country string) (string, error) {
	if country == "" {
		country = "RU"
	}
	number, err := libphonenumber.Parse(phoneNumber, country)
	if err != nil {
		return "", err
	}
	if !libphonenumber.IsValidNumber(number) {
		return "", errors.New("Phone number isn't valid")
	}
	return libphonenumber.Format(number, libphonenumber.E164), nil
}
