package models

import (
	"fmt"

	"proto/payment"

	"github.com/durango/go-credit-card"
	"github.com/jinzhu/gorm"
)

// Payment defines payment order info
type Payment struct {
	gorm.Model

	LeadID uint64

	// p2p params
	ShopCardNumber     string
	CustomerCardNumber string
	Amount             uint64
}

// Session once-used pay sess
type Session struct {
	gorm.Model

	PaymentID uint
	Amount    uint64
	IP        string

	// I wonder why payture wants 2 unique ids;
	UniqueID   string `gorm:"index"` // this one is used to check pay status
	ExternalID string `gorm:"index"` // this one is used by client
}

// Repo is mockable payment repository
type Repo interface {

	// pay part
	GetPayByID(uint) (*Payment, error)
	CreatePay(*Payment) error
	SavePay(*Payment) error

	// sess part
	CreateSess(*Session) error
}

// RepoImpl is implementation that uses gorm
type RepoImpl struct {
	DB *gorm.DB
}

// CreateSess to DB
func (r *RepoImpl) CreateSess(s *Session) error {
	return r.DB.Create(s).Error
}

// GetPayByID returns payment by ID
func (r *RepoImpl) GetPayByID(id uint) (*Payment, error) {
	var result Payment
	err := r.DB.Where("id = ?", id).Find(&result).Error
	return &result, err
}

// CreatePay creates payment
func (r *RepoImpl) CreatePay(p *Payment) error {
	return r.DB.Create(p).Error
}

// SavePay saves payment
func (r *RepoImpl) SavePay(p *Payment) error {
	return r.DB.Save(p).Error
}

// NewPayment generates a payment infocard by request
func NewPayment(r *payment.CreateOrderRequest) (*Payment, error) {

	// Check credit cards first
	if !validate(r.ShopCardNumber) || !validate(r.CustomerCardNumber) {
		return nil, fmt.Errorf("Invalid credit card (Luhn failed)")
	}

	if r.LeadId <= 0 {
		return nil, fmt.Errorf("Empty lead ID")
	}

	pay := &Payment{
		// Bank cards
		ShopCardNumber:     r.ShopCardNumber,
		CustomerCardNumber: r.CustomerCardNumber,

		// Core LeadID
		LeadID: r.LeadId,
	}

	switch r.Currency {
	case payment.Currency_RUB:

		// must convert to cops (1/100 of rub)
		pay.Amount = r.Amount * 100

	default:
		// unknown currency! panic
		return nil, fmt.Errorf("Unsupported currency %v (%v)", r.Currency, payment.Currency_name[int32(r.Currency)])
	}

	return pay, nil

}

// decorator to call CC check
func validate(cc string) bool {
	card := creditcard.Card{Number: cc}
	return card.ValidateNumber()
}
