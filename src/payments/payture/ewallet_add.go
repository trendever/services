package payture

import (
	"fmt"
	"proto/payment"
)

// Add new card (redirect user)
func (c *Ewallet) Add(info *payment.UserInfo) (string, error) {

	resp, err := c.vwInit(sessionTypeAdd, c.KeyAdd, info, nil)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("Error (%v) while AddCard init", resp.ErrCode)
	}

	return fmt.Sprintf("%v%v?SessionId=%v", c.URL, vwAddPath, resp.SessionID), nil
}

// GetCards checks given session status
func (c *Ewallet) GetCards(info *payment.UserInfo) (err error) {
	return nil
}
