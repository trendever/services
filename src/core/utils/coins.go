package utils

import (
	"core/api"
	"errors"
	"proto/trendcoin"
	"utils/nats"
	"utils/rpc"
)

func PerformTransactions(transactions ...*trendcoin.TransactionData) ([]uint64, error) {
	// @TODO add local checks
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := api.TrendcoinServiceClient.MakeTransactions(
		ctx,
		&trendcoin.MakeTransactionsRequest{Transactions: transactions},
	)
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, errors.New(res.Error)
	}
	return res.TransactionIds, nil
}

// Sends transactions request via nats.
// There is no way to know if request failed after msg was acknowledged by stan server,
// so it's bad idea to use it to write-off coins without allowed credit.
// Idempotency key is required in all transactions
func PostTransactions(transactions ...*trendcoin.TransactionData) error {
	for _, tx := range transactions {
		if tx.IdempotencyKey == "" {
			return errors.New("all transactions should have idempotency key")
		}
	}
	return nats.StanPublish("coins.makeTransactions", &trendcoin.MakeTransactionsRequest{
		Transactions: transactions,
	})
}
