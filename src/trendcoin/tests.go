package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"utils/db"
	"utils/log"
)

var tests = []struct {
	Name         string
	Transactions TransactionsSlice
	Expection    error
}{
	{
		Name: "init first account",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    1,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: nil,
	},
	{
		Name: "init second account",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    2,
				Amount:         100,
				Reason:         "reasonable",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: nil,
	},
	{
		Name: "already performed",
		Transactions: TransactionsSlice{
			&Transaction{
				// first transaction should be in log already
				ID:             1,
				Source:         0,
				Destination:    1,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: AlreadyPerformedError,
	},
	{
		Name: "zero amount",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    1,
				Amount:         0,
				Reason:         "reasonable",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: ZeroAmountError,
	},
	{
		Name: "empty reason",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    1,
				Amount:         10,
				Reason:         "",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: EmptyReasonError,
	},
	{
		Name: "two empty sides",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    0,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: true,
				AllowCredit:    false,
			},
		},
		Expection: TwoEmptySidesError,
	},
	{
		Name: "empty side",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         0,
				Destination:    1,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: EmptySideError,
	},
	{
		Name: "invalid source",
		Transactions: TransactionsSlice{
			&Transaction{
				// this account should not exists
				Source:         100500,
				Destination:    1,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: InvalidSourceError,
	},
	{
		Name: "identical sides",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         1,
				Destination:    1,
				Amount:         10,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: IdenticalSidesError,
	},
	{
		Name: "unexpected credit",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:      1,
				Destination: 2,
				// definitely negative balance in result
				Amount:         100500,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: CreditNotAllowedError,
	},
	{
		Name: "expected credit",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:      1,
				Destination: 2,
				// definitely negative balance in result
				Amount:         100500,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    true,
			},
		},
		Expection: nil,
	},
	{
		Name: "normal transfer",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:      2,
				Destination: 1,
				// transfer it back
				Amount:         100500,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: nil,
	},
	{
		Name: "normal multiple transactions",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         1,
				Destination:    2,
				Amount:         5,
				Reason:         "ping",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
			&Transaction{
				Source:         2,
				Destination:    1,
				Amount:         5,
				Reason:         "pong",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
			&Transaction{
				Source:         1,
				Destination:    2,
				Amount:         5,
				Reason:         "ping",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
			&Transaction{
				Source:         2,
				Destination:    1,
				Amount:         5,
				Reason:         "pong",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: nil,
	},
	{
		Name: "multiple transactions with trouble in second",
		Transactions: TransactionsSlice{
			&Transaction{
				Source: 1,
				// will create new account
				Destination:    3,
				Amount:         5,
				Reason:         "ping",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
			// should fail with EmptySideError
			&Transaction{
				Source:         3,
				Destination:    0,
				Amount:         5,
				Reason:         "miss!",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
			// may fail with CreditNotAllowedError, but should not be performed any way
			&Transaction{
				Source:         3,
				Destination:    1,
				Amount:         100500,
				Reason:         "wat",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: EmptySideError,
	},
	{
		Name: "rollback test(continuing previous)",
		Transactions: TransactionsSlice{
			&Transaction{
				// this account was created in previous test, but had to be rollbacked
				Source:         3,
				Destination:    1,
				Amount:         5,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
			},
		},
		Expection: InvalidSourceError,
	},
	{
		Name: "idempotency ckeck",
		Transactions: TransactionsSlice{
			&Transaction{
				Source:         2,
				Destination:    1,
				Amount:         1,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
				IdempotencyKey: "key",
			},
			&Transaction{
				Source:         2,
				Destination:    1,
				Amount:         1,
				Reason:         "reasonable",
				AllowEmptySide: false,
				AllowCredit:    false,
				// same as before
				IdempotencyKey: "key",
			},
		},
		Expection: IdempotencyCheckError,
	},
}

func runTests(configPath string) {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("filed to read test config: %v", err))
	}
	err = yaml.Unmarshal(file, &settings)
	if err != nil {
		log.Errorf("filed to parse config: %v", err)
	}
	db.Init(&settings.DB)
	Migrate(true)
	for _, test := range tests {
		result := test.Transactions.Perform()
		if result != test.Expection {
			log.Fatal(fmt.Errorf(
				"test '%v' failed:\n\t expected '%v',\n\t but result is '%v'!",
				test.Name, test.Expection, result,
			))
		} else {
			log.Info("test '%v' passed", test.Name)
			checkIntegrity()
		}
	}
}

// checks whether balances of all accounts match transactions log
func checkIntegrity() {
	var results []struct {
		UserID  uint64
		Balance int64
		Diff    int64
	}
	err := db.New().Raw(`
SELECT acc.user_id, acc.balance, COALESCE(p.sum, 0) - COALESCE(m.sum, 0) AS diff
FROM accounts acc
LEFT JOIN (SELECT SUM(amount), source FROM transactions GROUP BY source) AS m
ON m.source = acc.user_id
LEFT JOIN (SELECT SUM(amount), destination FROM transactions GROUP BY destination) AS p
ON p.destination = acc.user_id
WHERE COALESCE(p.sum, 0) - COALESCE(m.sum, 0) != acc.balance
	`).Scan(&results).Error
	if err != nil {
		log.Fatal(fmt.Errorf("failed to check integrity: %v", err))
	}
	if len(results) != 0 {
		log.Info("integrity test failed for following accounts:")
		for _, acc := range results {
			log.Info("user_id: %v,\tbalance: %v,\tdiff: %v", acc.UserID, acc.Balance, acc.Diff)
		}
		log.Fatal(errors.New("abrot"))
	}
}
