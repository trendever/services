package payture

import (
	"fmt"
	"payments/models"
	"proto/payment"
	"utils/log"

	"github.com/pborman/uuid"
)

// Buy request
func (c *Ewallet) Buy(pay *models.Payment, info *payment.UserInfo) (*models.Session, error) {

	pd := &payDef{
		orderID: uuid.New(),
	}

	session := models.Session{
		PaymentID:   pay.ID,
		IP:          info.Ip,
		GatewayType: vwGwType,
	}

	switch payment.Currency(pay.Currency) {
	case payment.Currency_RUB:
		// must convert to cops (1/100 of rub)
		pd.amount = pay.Amount * 100
	case payment.Currency_COP:
		pd.amount = pay.Amount
	default:
		// unknown currency! panic
		session.FailureReason = "Bad currency"
		log.Errorf("Unsupported currency %v (%v)", pay.Currency, payment.Currency_name[pay.Currency])
		return &session, nil
	}

	if pay.CardID != "" {
		pd.cardID = pay.CardID
	} else { // find first active card
		cards, err := c.GetCards(info)
		if err != nil {
			session.FailureReason = fmt.Sprintf("Network error while fetching cards: %v", err)
			return &session, nil
		}
		card, err := firstActive(cards)
		if err != nil {
			session.FailureReason = fmt.Sprintf("No active card error: %v", err)
			return &session, nil
		}
		pd.cardID = card.Id
	}

	res, err := c.vwPay(sessionTypePay, c.KeyPay, info, pd)
	if err != nil {
		session.FailureReason = fmt.Sprintf("Network error: %v", err)
		return &session, nil
	}

	if !res.Success {
		log.Errorf("Error (%v) while AddCard init (pay id=%v)", res.ErrCode, pay.ID)
		session.FailureReason = res.ErrCode
	}

	session.ExternalID = res.MerchantOrderID
	session.UniqueID = res.SessionID
	session.Amount = res.Amount
	session.Success = res.Success

	return &session, nil
}

// Redirect returns client-redirectable redirect link
func (c *Ewallet) Redirect(sess *models.Session) string {
	return ""
}

// CheckStatus checks given session status
func (c *Ewallet) CheckStatus(sess *models.Session) (finished bool, err error) {
	// no need to do any futher checks; just mark as finished
	sess.Finished = true
	return true, nil
}