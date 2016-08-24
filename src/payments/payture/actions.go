package payture

import (
	"encoding/xml"
	"fmt"
	"payments/config"
	"payments/models"

	"github.com/pborman/uuid"
)

const (
	initMethod      = "Init"
	payStatusMethod = "PayStatus"
)

const successState = "Charged"

const gatewayType = "payture"

// GatewayType for this pkg
func (c *Client) GatewayType() string {
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
func (c *Client) Buy(pay *models.Payment, ipAddr string) (*models.Session, error) {

	var res initResponse

	uniqueID := uuid.New()

	err := c.xmlRequest(initMethod, &res, map[string]string{
		"SessionType": "Pay",
		"OrderID":     uniqueID,
		"Amount":      fmt.Sprintf("%v", pay.Amount),
		"IP":          ipAddr,

		// callback URL; seems not to work in sandbox mode
		"Url": config.Get().HTTP.Public + "?orderid={orderid}&success={success}",

		// template fields
		"Product": fmt.Sprintf("#%d", pay.LeadID),
		"Total":   fmt.Sprintf("%v", pay.Amount),
	}, nil)

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
		IP:          ipAddr,
		GatewayType: gatewayType,
	}, nil
}

// Redirect returns client-redirectable redirect link
func (c *Client) Redirect(sess *models.Session) string {
	return fmt.Sprintf("%v/apim/Pay?SessionId=%v", c.URL, sess.ExternalID)
}

// CheckStatus checks given session status
func (c *Client) CheckStatus(sess *models.Session) (finished bool, err error) {
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
	sess.Finished = (res.OrderID != "") && (res.Success || res.ErrCode != "NONE")
	sess.State = res.State

	finished = sess.Finished

	return
}
