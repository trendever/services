package payture

import (
	"errors"
	"fmt"
	"payments/models"
	"proto/payment"

	"github.com/pborman/uuid"
)

// Buy request
func (c *Ewallet) Buy(pay *models.Payment, info *payment.UserInfo, async bool) (*models.Session, error) {

	pd := &payDef{
		orderID: uuid.New(),
	}

	session := models.Session{
		PaymentID:   pay.ID,
		Payment:     pay,
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
		return &session, fmt.Errorf("Unsupported currency %v (%v)", pay.Currency, payment.Currency_name[pay.Currency])
	}

	if pay.CardID != "" {
		pd.cardID = pay.CardID
	} else if async { // find first active card
		cards, err := c.GetCards(info)
		if err != nil {
			session.FailureReason = fmt.Sprintf("Network error while fetching cards: %v", err)
			return &session, errors.New(session.FailureReason)
		}
		card, err := firstActive(cards)
		if err != nil {
			session.FailureReason = fmt.Sprintf("No active card error: %v", err)
			return &session, errors.New(session.FailureReason)
		}
		pd.cardID = card.Id
	}

	if async {
		res, err := c.vwPayAsync(info, pd)
		if err != nil {
			session.FailureReason = fmt.Sprintf("Network error: %v", err)
			return &session, errors.New(session.FailureReason)
		}

		session.ExternalID = res.MerchantOrderID
		session.UniqueID = res.SessionID
		session.Amount = res.Amount
		session.Success = res.Success
		session.Finished = true

		if !res.Success {
			session.FailureReason = res.ErrCode
			return &session, fmt.Errorf("Error (%v) while Pay init (pay id=%v)", res.ErrCode, pay.ID)
		}
	} else {
		res, err := c.vwPay(info, pd)
		if err != nil {
			session.FailureReason = fmt.Sprintf("Network error: %v", err)
			return &session, errors.New(session.FailureReason)
		}

		session.UniqueID = res.OrderID
		session.ExternalID = res.SessionID
		session.Success = res.Success

		if !res.Success {
			session.FailureReason = res.ErrCode
			return &session, fmt.Errorf("Error (%v) while Pay init (pay id=%v)", res.ErrCode, pay.ID)
		}

	}

	return &session, nil
}

// Redirect returns client-redirectable redirect link
func (c *Ewallet) Redirect(sess *models.Session) string {
	return fmt.Sprintf("%v%v?SessionId=%v", c.URL, vwPayPath, sess.ExternalID)
}

// CheckStatus checks given session status
func (c *Ewallet) CheckStatus(sess *models.Session) (bool, error) {

	res, err := c.vwPayStatus(sess.UniqueID)
	if err != nil {
		return false, err
	}

	sess.Success = (res.State == successState)
	sess.Finished = (res.ErrCode == timeoutError || res.ErrCode == notfoundError || sess.Success)
	sess.State = res.State
	sess.FailureReason = res.ErrCode

	return sess.Finished, nil
}
