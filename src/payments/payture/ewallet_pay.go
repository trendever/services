package payture

import (
	"fmt"
	"payments/models"
	"proto/payment"
)

// Buy request
func (c *Ewallet) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

// Redirect returns client-redirectable redirect link
func (c *Ewallet) Redirect(sess *models.Session) string {
	return "" //fmt.Errorf("Not yet implemented")
}

// CheckStatus checks given session status
func (c *Ewallet) CheckStatus(sess *models.Session) (finished bool, err error) {
	return false, fmt.Errorf("Not yet coded")
}
