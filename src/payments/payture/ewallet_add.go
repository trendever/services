package payture

import (
	"fmt"
	"payments/models"
	"proto/payment"
)

const addGatewayType = "payture_ewallet_add"

// GatewayType for this pkg
func (c *EwalletAdd) GatewayType() string {
	return addGatewayType
}

// Buy request
func (c *EwalletAdd) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {

	if pay != nil {
		return nil, fmt.Errorf("Adding card do not require payment")
	}

	resp, err := c.Ewallet.vwInit(sessionTypeAdd, c.KeyAdd, info, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Error (%v) while AddCard init", resp.ErrorCode)
	}

	return &models.Session{
		ExternalID:  res.SessionID,
		UniqueID:    res.SessionID,
		IP:          info.Ip,
		GatewayType: addGatewayType,
	}, nil
}

// Redirect returns client-redirectable redirect link
func (c *EwalletAdd) Redirect(sess *models.Session) string {
	return "" //fmt.Errorf("Not yet implemented")
}

// CheckStatus checks given session status
func (c *EwalletAdd) CheckStatus(sess *models.Session) (finished bool, err error) {
	return false, fmt.Errorf("Not yet coded")
}
