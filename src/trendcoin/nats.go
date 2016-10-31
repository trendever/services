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
	UserID uint64 `json:"user_id"`
	Amount uint64 `json:"amount"`
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
			Subject:        "payments.notifications",
			Group:          "trendcoin",
			DurableName:    "trendcoin",
			AckTimeout:     time.Second * 15,
			DecodedHandler: s.NatsTransactions,
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
	// in case of external(db) error we want receive this request again later
	externalErr := true
	if res.Error != "" {
		for _, cur := range LogicalErrors {
			if res.Error == cur.Error() {
				externalErr = false
				break
			}
		}
	}
	return !externalErr
}

func (s *TrendcoinServer) NatsRefill(in *payment.PaymentNotification) (acknowledged bool) {
	if in.Data.ServiceName != PaymentName {
		return true
	}
	log.Debug("got payment notification with data %v", in.Data.ServiceData)
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
	return s.NatsTransactions(&trendcoin.MakeTransactionsRequest{
		Transactions: []*trendcoin.TransactionData{{
			Destination:    data.UserID,
			Amount:         data.Amount,
			AllowEmptySide: true,
			Reason:         fmt.Sprintf("refilled with payment @%v", in.Id),
			IdempotencyKey: fmt.Sprintf("payment %v", in.Id),
		}},
	})
}
