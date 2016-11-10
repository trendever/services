package payture

import (
	"encoding/xml"
	"fmt"
	"payments/config"
	"payments/models"
	"proto/payment"

	"github.com/pborman/uuid"
)

const (
	initMethod      = "Init"
	payStatusMethod = "PayStatus"
)

const (
	successState = "Charged"
	timeoutError = "ORDER_TIME_OUT"
)

const gatewayType = "payture"

// GatewayType for this pkg
func (c *InPay) GatewayType() string {
	return gatewayType
}

type initResponse struct {
	XMLName   xml.Name `xml:"Init"`
	Success   bool     `xml:"Success,attr"`
	OrderID   string   `xml:"OrderId,attr"`
	Amount    uint64   `xml:"Amount,attr"`
	SessionID string   `xml:"SessionId,attr"`
	ErrCode   string   `xml:"ErrCode,attr"`
}

type payStatusResponse struct {
	XMLName xml.Name `xml:"PayStatus"`
	Success bool     `xml:"Success,attr"`
	OrderID string   `xml:"OrderId,attr"`
	Amount  uint64   `xml:"Amount,attr"`
	State   string   `xml:"State,attr"`
	ErrCode string   `xml:"ErrCode,attr"`
}

// Buy request
func (c *InPay) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {

	if info == nil || pay == nil {
		return nil, fmt.Errorf("payments/payture: got nil userInfo or pay")
	}

	var res initResponse

	uniqueID := uuid.New()

	var amount uint64

	switch payment.Currency(pay.Currency) {
	case payment.Currency_RUB:
		// must convert to cops (1/100 of rub)
		amount = pay.Amount * 100
	case payment.Currency_COP:
		amount = pay.Amount
	default:
		// unknown currency! panic
		return nil, fmt.Errorf("Unsupported currency %v (%v)", pay.Currency, payment.Currency_name[pay.Currency])
	}

	request := map[string]string{
		"SessionType": "Pay",
		"OrderID":     uniqueID,
		"Amount":      fmt.Sprintf("%v", amount),
		"IP":          info.Ip,

		// callback URL; seems not to work in sandbox mode
		// or work at all? file support to change it
		"Url": config.Get().HTTP.Public + "?orderid={orderid}&success={success}",

		// template fields
		"CardTo": pay.ShopCardNumber,
		// convert back to rubs
		"Total": fmt.Sprintf("%d", amount/100),
	}

	err := c.xmlRequest(initMethod, &res, request, nil)

	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, fmt.Errorf("Unsuccessfull payment init (orderID %v); errorCode: %v", uniqueID, res.ErrCode)
	}

	return &models.Session{
		PaymentID:   pay.ID,
		ExternalID:  res.SessionID,
		UniqueID:    uniqueID,
		Amount:      res.Amount,
		IP:          info.Ip,
		GatewayType: gatewayType,
	}, nil
}

// Redirect returns client-redirectable redirect link
func (c *InPay) Redirect(sess *models.Session) string {
	return fmt.Sprintf("%v/apim/Pay?SessionId=%v", c.URL, sess.ExternalID)
}

// CheckStatus checks given session status
func (c *InPay) CheckStatus(sess *models.Session) (finished bool, err error) {
	var res payStatusResponse

	err = c.xmlRequest(payStatusMethod, &res, nil, map[string]string{
		"OrderId": sess.UniqueID,
	})
	if err != nil {
		return
	}

	if !res.Success && res.OrderID == "" {
		err = fmt.Errorf("Unsuccessfull PayStatus: errCode=%v", res.ErrCode)
		return
	}

	sess.Success = (res.State == successState)
	sess.Finished = (res.ErrCode == timeoutError)
	sess.State = res.State

	finished = sess.Finished

	return
}
