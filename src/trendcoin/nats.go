package main

import (
	"encoding/json"
	"fmt"
	"proto/payment"
	"proto/trendcoin"
	"time"
	"utils/log"
	"utils/nats"
)

const PaymentName = "coins_refill"

type PaymentData struct {
	UserID     uint64 `json:"user_id"`
	Amount     uint64 `json:"amount"`
	AutoRefill bool   `json:"auto_refill"`
}

func (s *TrendcoinServer) subscribe() {
	nats.StanSubscribe(
		&nats.StanSubscription{
			Subject:        "coins.makeTransactions",
			Group:          "trendcoin",
			DurableName:    "trendcoin",
			AckTimeout:     time.Second * 15,
			DecodedHandler: s.NatsTransactions,
		},
		&nats.StanSubscription{
			Subject:        "payments.event",
			Group:          "trendcoin",
			DurableName:    "trendcoin",
			AckTimeout:     time.Second * 15,
			DecodedHandler: s.NatsRefill,
		},
	)
}

func (s *TrendcoinServer) NatsTransactions(in *trendcoin.MakeTransactionsRequest) (acknowledged bool) {
	log.Debug("got transactions request via nats: %+v", in)
	for _, tx := range in.Transactions {
		if tx.IdempotencyKey == "" {
			log.Errorf("nats transaction request %+v without IdempotencyKey ignored", in)
			return true
		}
	}
	res, _ := s.MakeTransactions(nil, in)
	if res.Error != "" {
		// in case of external(db) error we want receive this request again later
		externalErr := true
		for _, cur := range LogicalErrors {
			if res.Error == cur.Error() {
				externalErr = false
				break
			}
		}
		return !externalErr
	}
	return true
}

func (s *TrendcoinServer) NatsRefill(in *payment.PaymentNotification) (acknowledged bool) {
	if in.Data.ServiceName != PaymentName {
		return true
	}
	log.Debug("got payment notification %+v", in)

	var data PaymentData
	err := json.Unmarshal([]byte(in.Data.ServiceData), &data)
	if err != nil {
		log.Errorf("failed to unmarshal data of payment %v: %v", in.Id, err)
		return true
	}
	if data.Amount == 0 || data.UserID == 0 {
		log.Errorf("invalid refill data %+v", data)
		return true
	}

	switch in.Event {
	case payment.Event_PaySuccess:
		return s.NatsTransactions(&trendcoin.MakeTransactionsRequest{
			Transactions: []*trendcoin.TransactionData{{
				Destination:    data.UserID,
				Amount:         data.Amount,
				AllowEmptySide: true,
				Reason:         fmt.Sprintf("refilled with payment @%v", in.Id),
				IdempotencyKey: fmt.Sprintf("payment %v", in.Id),
			}},
			IsAutorefill: data.AutoRefill,
		})

	case payment.Event_PayFailed:
		if data.AutoRefill {
			err := nats.StanPublish(NotifyTopic, &trendcoin.BalanceNotify{
				UserId:     data.UserID,
				Autorefill: true,
				Failed:     true,
			})
			if err != nil {
				log.Errorf("failed to notify about failed autorefill: %v", err)
				return false
			}
		}
	}
	return true
}
