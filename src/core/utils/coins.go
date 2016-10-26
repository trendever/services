package utils

import (
	"core/api"
	"errors"
	"proto/trendcoin"
	"utils/rpc"
)

func PerformTransactions(transactions ...*trendcoin.TransactionData) error {
	// @TODO add local checks
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := api.TrendcoinServiceClient.MakeTransactions(
		ctx,
		&trendcoin.MakeTransactionsRequest{Transactions: transactions},
	)
	if err != nil {
		return err
	}
	if res.Error != "" {
		return errors.New(res.Error)
	}
	return nil
}
