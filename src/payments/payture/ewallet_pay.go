package payture

import (
	"fmt"
	"payments/models"
	"proto/payment"

	"github.com/pborman/uuid"
)

// Buy request
func (c *Ewallet) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {

	pd := &payDef{
		orderID: uuid.New(),
	}

	switch payment.Currency(pay.Currency) {
	case payment.Currency_RUB:
		// must convert to cops (1/100 of rub)
		pd.amount = pay.Amount * 100
	case payment.Currency_COP:
		pd.amount = pay.Amount
	default:
		// unknown currency! panic
		return nil, fmt.Errorf("Unsupported currency %v (%v)", pay.Currency, payment.Currency_name[pay.Currency])
	}

	if pay.CardID != "" {
		pd.cardID = pay.CardID
	} else { // find first active card
		cards, err := c.GetCards(info)
		if err != nil {
			return nil, err
		}
		card, err := firstActive(cards)
		if err != nil {
			return nil, err
		}
		pd.cardID = card.Id
	}

	res, err := c.vwPay(sessionTypePay, c.KeyPay, info, pd)
	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, fmt.Errorf("Error (%v) while AddCard init", res.ErrCode)
	}
	return &models.Session{
		PaymentID:   pay.ID,
		ExternalID:  res.MerchantOrderID,
		UniqueID:    res.SessionID,
		Amount:      res.Amount,
		IP:          info.Ip,
		GatewayType: vwGwType,
	}, nil
}

// Redirect returns client-redirectable redirect link
func (c *Ewallet) Redirect(sess *models.Session) string {
	return "" //fmt.Errorf("Not yet implemented")
}

// CheckStatus checks given session status
func (c *Ewallet) CheckStatus(sess *models.Session) (finished bool, err error) {
	return true, fmt.Errorf("Not yet coded")
}
