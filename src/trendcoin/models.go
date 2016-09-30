package main

import (
	proto "proto/trendcoin"
	"time"
	"utils/db"
	"utils/log"
)

var dbModels = []interface{}{
	&Transaction{},
	&Account{},
}

type Account struct {
	UserID  uint64 `gorm:"primary_key"`
	Balance int64
}

type Transaction struct {
	ID             uint64    `gorm:"primary_key"`
	CreatedAt      time.Time `gorm:"index"`
	Source         uint64    `gorm:"index" sql:"default:NULL"`
	Destination    uint64    `gorm:"index" sql:"default:NULL"`
	Amount         uint64
	Reason         string
	AllowCredit    bool
	AllowEmptySide bool
}

func (t Transaction) Encode() *proto.Transaction {
	return &proto.Transaction{
		Id:        t.ID,
		CreatedAt: t.CreatedAt.UnixNano(),
		Data: &proto.TransactionData{
			Source:         t.Source,
			Destination:    t.Destination,
			Amount:         t.Amount,
			Reason:         t.Reason,
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
	db.Model(&Transaction{}).AddForeignKey("source", "accounts(user_id)", "RESTRICT", "RESTRICT")
	db.Model(&Transaction{}).AddForeignKey("destination", "accounts(user_id)", "RESTRICT", "RESTRICT")
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
