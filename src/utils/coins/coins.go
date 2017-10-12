package coins

import (
	"common/log"
	"errors"
	"fmt"
	"proto/trendcoin"
	"utils/nats"
	"utils/rpc"
)

var (
	MissingIdempotencyKey = errors.New("all transactions should have idempotency key")
	InsufficientFunds     = errors.New("insufficient funds")
	ServiceError          = errors.New("temporarily unable to write-off coins")
	CallbackFailed        = errors.New("call back return non nil error")
	RefundError           = errors.New("failed to refund coins after callback error")
)

var grpcCli trendcoin.TrendcoinServiceClient

func SetGRPCCli(cli trendcoin.TrendcoinServiceClient) {
	grpcCli = cli
}

// Performs transactions via grpc, its client should be set with SetGRPCCli() before call
func PerformTransactions(transactions ...*trendcoin.TransactionData) ([]uint64, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := grpcCli.MakeTransactions(
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

// Sends transactions request via stan(connection should be inited by caller).
// There is no way to know if request failed after msg was acknowledged by stan server,
// so it's bad idea to use it to write-off coins without allowed credit.
// Idempotency key is required in all transactions
func PostTransactions(transactions ...*trendcoin.TransactionData) error {
	for _, tx := range transactions {
		if tx.IdempotencyKey == "" {
			return MissingIdempotencyKey
		}
	}
	return nats.StanPublish("coins.makeTransactions", &trendcoin.MakeTransactionsRequest{
		Transactions: transactions,
	})
}

// Tries to write-off coins, calls callback if success, refunds coins if cb returns an error.
// Requires both grpcCli set and active stan connection.
func CheckWriteOff(userID, amount uint64, reason string, allowCredit bool, cb func() error) error {
	txIDs, err := PerformTransactions(&trendcoin.TransactionData{
		Source:         userID,
		Amount:         amount,
		AllowEmptySide: true,
		AllowCredit:    allowCredit,
		Reason:         reason,
	})
	if err != nil {
		if err.Error() == "Invalid source account" || err.Error() == "Credit is not allowed for this transaction" {
			return InsufficientFunds
		}
		log.Errorf("failed to perform transactions: %v", err)
		return ServiceError
	}

	err = cb()
	// here comes troubles
	if err != nil {
		log.Errorf("callback returns error: %v", err)
		refundErr := PostTransactions(&trendcoin.TransactionData{
			Destination:    userID,
			Amount:         amount,
			AllowEmptySide: true,
			Reason:         fmt.Sprintf("refund #%v after internal error", txIDs[0]),
			IdempotencyKey: fmt.Sprintf("#%v refund", txIDs[0]),
		})
		if refundErr != nil {
			// well... things going really bad
			// @TODO we need extra error log level, it's really critical
			log.Errorf("failed to refund %v coins to %v: %v!", amount, userID, refundErr)
			return RefundError
		}
		return CallbackFailed
	}
	return nil
}
