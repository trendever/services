package main

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"utils/db"
	"utils/log"
)

var (
	AlreadyPerformedError = errors.New("Transaction has already performed")
	ZeroAmountError       = errors.New("Transaction amount should not be zero")
	EmptyReasonError      = errors.New("Transaction reason should not be empty")
	TwoEmptySidesError    = errors.New("Transaction should not have two empty sides")
	EmptySideError        = errors.New("Emply side is not allowed for this transaction")
	InvalidSourceError    = errors.New("Invalid source account")
	IdenticalSidesError   = errors.New("Transaction should not have identical source and destination")
	CreditNotAllowedError = errors.New("Credit is not allowed for this transaction")
)

// performs basic transaction validate
func (t *Transaction) Validate() error {
	if t.ID != 0 {
		return AlreadyPerformedError
	}
	if t.Reason == "" {
		return EmptyReasonError
	}
	if t.Amount == 0 {
		return ZeroAmountError
	}
	if t.Source == 0 && t.Destination == 0 {
		return TwoEmptySidesError
	}
	if t.Source == t.Destination {
		return IdenticalSidesError
	}
	if !t.AllowEmptySide && (t.Source == 0 || t.Destination == 0) {
		return EmptySideError
	}
	return nil
}

// db transaction should be already started on higher level
func (t *Transaction) Perform(tx *gorm.DB) error {
	if err := t.Validate(); err != nil {
		return err
	}
	if t.Source != 0 {
		source := Account{UserID: t.Source}
		res := tx.Find(&source)
		// source account should exist
		// @CHECK or not? if credit is allowed it may be fine
		if res.RecordNotFound() {
			return InvalidSourceError
		}
		if res.Error != nil {
			return fmt.Errorf("failed to load source account: %v", res.Error)
		}
		if !t.AllowCredit && source.Balance < int64(t.Amount) {
			return CreditNotAllowedError
		}
		source.Balance -= int64(t.Amount)
		err := tx.Save(&source).Error
		if err != nil {
			return fmt.Errorf("failed to save source account: %v", res.Error)
		}
	}
	if t.Destination != 0 {
		destination := Account{UserID: t.Destination}
		res := tx.Find(&destination)
		if res.Error != nil && !res.RecordNotFound() {
			return fmt.Errorf("failed to load destination account: %v", res.Error)
		}
		destination.Balance += int64(t.Amount)
		res = tx.Save(&destination)
		if res.Error != nil {
			return fmt.Errorf("failed to save destination account: %v", res.Error)
		}
	}
	err := tx.Save(t).Error
	if err != nil {
		return fmt.Errorf("failed to save transaction into log: %v", err)
	}

	return nil
}

func (slice TransactionsSlice) Perform() error {
	tx := db.NewTransaction()
	for _, trans := range slice {
		err := trans.Perform(tx)
		if err == nil {
			continue
		}
		log.Errorf("while performing transaction %+v: %v", trans, err)
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
