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
func (c *Ewallet) GetCards(info *payment.UserInfo) ([]*payment.Card, error) {
	resp, err := c.vwCards(info)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Error (%v) while GetCards", resp.ErrCode)
	}

	result := make([]*payment.Card, len(resp.Items))
	for i, item := range resp.Items {
		result[i] = &payment.Card{
			Id:     item.CardID,
			Name:   item.CardName,
			Active: item.Status == "IsActive",
		}
	}

	return result, nil
}

func firstActive(cards []*payment.Card) (*payment.Card, error) {
	for _, card := range cards {
		if card.Active {
			return card, nil
		}
	}

	return nil, fmt.Errorf("Found no active card")
}

// DelCard just deletes card
func (c *Ewallet) DelCard(cardID string, info *payment.UserInfo) error {
	resp, err := c.vwDelCard(cardID, info)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("Error (%v) while GetCards", resp.ErrCode)
	}

	return nil
}
