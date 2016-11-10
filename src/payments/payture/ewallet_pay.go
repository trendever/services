package payture

import (
	"fmt"
	"payments/models"
	"proto/payment"
)

const payGatewayType = "payture_ewallet_pay"

// GatewayType for this pkg
func (c *EwalletPay) GatewayType() string {
	return payGatewayType
}

// Buy request
func (c *EwalletPay) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

// Redirect returns client-redirectable redirect link
func (c *EwalletPay) Redirect(sess *models.Session) string {
	return "" //fmt.Errorf("Not yet implemented")
}

// CheckStatus checks given session status
func (c *EwalletPay) CheckStatus(sess *models.Session) (finished bool, err error) {
	return false, fmt.Errorf("Not yet coded")
}
