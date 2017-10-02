package main

import (
	"common/db"
	"common/log"
	proto "proto/trendcoin"
	"time"
)

var dbModels = []interface{}{
	&Transaction{},
	&Account{},
}

type Account struct {
	UserID  uint64 `gorm:"primary_key"`
	Balance int64
}

func (a Account) TableName() string {
	return "coins_balance"
}

type Transaction struct {
	ID             uint64    `gorm:"primary_key"`
	CreatedAt      time.Time `gorm:"index"`
	Source         uint64    `gorm:"index" sql:"default:NULL"`
	Destination    uint64    `gorm:"index" sql:"default:NULL"`
	Amount         uint64
	Reason         string
	IdempotencyKey string `gorm:"unique" sql:"default:NULL"`
	AllowCredit    bool
	AllowEmptySide bool
}

func (t Transaction) Encode() *proto.Transaction {
	return &proto.Transaction{
		Id:        t.ID,
		CreatedAt: t.CreatedAt.Unix(),
		Data: &proto.TransactionData{
			Source:         t.Source,
			Destination:    t.Destination,
			Amount:         t.Amount,
			Reason:         t.Reason,
			IdempotencyKey: t.IdempotencyKey,
			AllowCredit:    t.AllowCredit,
			AllowEmptySide: t.AllowEmptySide,
		},
	}
}

func DecodeTransactionData(d *proto.TransactionData) *Transaction {
	return &Transaction{
		Source:         d.Source,
		Destination:    d.Destination,
		Amount:         d.Amount,
		Reason:         d.Reason,
		IdempotencyKey: d.IdempotencyKey,
		AllowCredit:    d.AllowCredit,
		AllowEmptySide: d.AllowEmptySide,
	}
}

type TransactionsSlice []*Transaction

func (slice TransactionsSlice) Encode() []*proto.Transaction {
	ret := make([]*proto.Transaction, 0, len(slice))
	for _, trans := range slice {
		ret = append(ret, trans.Encode())
	}
	return ret
}

func Migrate(drop bool) {
	db := db.New()

	// @TODO one-time drop, remove it after some revisions
	var count uint
	db.New().Table("information_schema.table_constraints").Where("table_name='transactions' and constraint_name = 'transactions_destination_accounts_user_id_foreign'").Count(&count)
	if count > 0 {
		drop = true
	}

	if drop {
		log.Warn("Droping tables")
		err := db.DropTableIfExists(dbModels...).Error
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := db.AutoMigrate(dbModels...).Error; err != nil {
		log.Fatal(err)
	}
	db.Model(&Transaction{}).AddForeignKey("source", "coins_balance(user_id)", "RESTRICT", "RESTRICT")
	db.Model(&Transaction{}).AddForeignKey("destination", "coins_balance(user_id)", "RESTRICT", "RESTRICT")
	db.Exec(`
CREATE OR REPLACE FUNCTION raise_modify_exception() RETURNS trigger AS
$x$
BEGIN
RAISE EXCEPTION 'Never modify transaction log';
END
$x$ LANGUAGE plpgsql;
	`)
	db.Exec(`
CREATE TRIGGER disable_changes BEFORE UPDATE OR DELETE
ON transactions
EXECUTE PROCEDURE raise_modify_exception();
	`)
}
