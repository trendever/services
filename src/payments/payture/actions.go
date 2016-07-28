package payture

import (
	"encoding/xml"
	"fmt"
	"github.com/pborman/uuid"
	"payments/models"
)

const initMethod = "Init"

type initResponse struct {
	XMLName   xml.Name `xml:"Init"`
	Success   bool     `xml:"Success,attr"`
	OrderID   string   `xml:"OrderId,attr"`
	Amount    uint64   `xml:"Amount,attr"`
	SessionID string   `xml:"SessionId,attr"`
	ErrCode   string   `xml:"ErrCode,attr"`
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

		// template fields
		"Product": "Trendever",
		"Total":   fmt.Sprintf("%v", pay.Amount),
	})

	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, fmt.Errorf("Unsuccessfull payment init (orderID %v); errorCode: %v", uniqueID, res.ErrCode)
	}

	return &models.Session{
		PaymentID:  pay.ID,
		ExternalID: res.SessionID,
		UniqueID:   uniqueID,
		Amount:     res.Amount,
		IP:         ipAddr,
	}, nil
}

// Redirect returns client-redirectable redirect link
func (c *Client) Redirect(sess *models.Session) string {
	return fmt.Sprintf("%v/apim/Pay?SessionId=%v", c.URL, sess.ExternalID)
}
