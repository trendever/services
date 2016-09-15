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

	LeadID         uint64 `gorm:"index"`
	Direction      int32
	ConversationID uint64
	UserID         uint64 // that's client id
	MessageID      uint64 // message id (in chat service) that contains payment button
	Cancelled      bool   `sql:"default:false"` // sender side can cancel payment

	// p2p params
	ShopCardNumber string
	Amount         uint64
	Currency       int32
}

// Session once-used pay sess
type Session struct {
	gorm.Model

	PaymentID uint
	Payment   *Payment

	Amount      uint64
	IP          string
	GatewayType string `gorm:"index"`

	State        string `gorm:"index"`
	Finished     bool   `gorm:"index" sql:"default:false"` // session can be finished, but unsuccessfully
	Success      bool   `gorm:"index" sql:"default:false"`
	ChatNotified bool   `gorm:"index" sql:"default:false"`

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
	CanCreateOrder(leadID uint) (bool, error)

	// sess part
	CreateSess(*Session) error
	GetSessByUID(string) (*Session, error)
	FinishedSessionsForPayID(pay uint) (int, error)
	SaveSess(*Session) error
	GetUnfinished(string) ([]Session, error)
}

// RepoImpl is implementation that uses gorm
type RepoImpl struct {
	DB *gorm.DB
}

// Gateway interface (1-step payment)
type Gateway interface {

	// create buying session
	Buy(sess *Payment, ipAddr string) (*Session, error)

	// get redirect URL for this session
	Redirect(*Session) string

	CheckStatus(*Session) (finished bool, err error)

	GatewayType() string
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

// GetSessByUID returns payment by ID
func (r *RepoImpl) GetSessByUID(uid string) (*Session, error) {
	var result Session
	err := r.DB.
		Where("unique_id = ?", uid).
		Preload("Payment").
		Find(&result).
		Error

	return &result, err
}

// FinishedSessionsForPayID returns num of successfull payments with given pay ID
func (r *RepoImpl) FinishedSessionsForPayID(payID uint) (int, error) {

	var count int
	err := r.DB.
		Model(&Session{}).
		Where("payment_id = ?", payID).
		Count(&count).
		Error

	return count, err
}

// GetUnfinished returns payment by ID
func (r *RepoImpl) GetUnfinished(gatewayType string) ([]Session, error) {
	var result []Session

	err := r.DB.
		Where("gateway_type = ?", gatewayType).
		Where("finished != TRUE or chat_notified != TRUE").
		Preload("Payment").
		Find(&result).
		Error

	return result, err
}

// SaveSess saves payment
func (r *RepoImpl) SaveSess(p *Session) error {
	return r.DB.Save(p).Error
}

// CreatePay creates payment
func (r *RepoImpl) CreatePay(p *Payment) error {
	return r.DB.Create(p).Error
}

// SavePay saves payment
func (r *RepoImpl) SavePay(p *Payment) error {
	return r.DB.Save(p).Error
}

// CanCreateOrder shows if you can create another order for this leadID
func (r *RepoImpl) CanCreateOrder(leadID uint) (bool, error) {

	var count int
	err := r.DB.
		Model(&Payment{}).
		Where("lead_id = ? and cancelled = FALSE", leadID).
		Joins("LEFT JOIN sessions as sess ON payments.id = sess.pay_id and sess.finished = FALSE").
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// NewPayment generates a payment infocard by request
func NewPayment(r *payment.CreateOrderRequest) (*Payment, error) {

	// Check credit cards first
	if !validate(r.ShopCardNumber) {
		return nil, fmt.Errorf("Invalid credit card (Luhn failed)")
	}

	if r.LeadId <= 0 {
		return nil, fmt.Errorf("Empty lead ID")
	}

	pay := &Payment{
		// Bank cards
		ShopCardNumber: r.ShopCardNumber,

		// Core LeadID
		LeadID:         r.LeadId,
		Direction:      int32(r.Direction),
		ConversationID: r.ConversationId,
		UserID:         r.UserId,
	}

	switch r.Currency {
	case payment.Currency_RUB:

		// must convert to cops (1/100 of rub)
		pay.Amount = r.Amount * 100
		pay.Currency = int32(payment.Currency_COP)

	case payment.Currency_COP:

		pay.Amount = r.Amount
		pay.Currency = int32(payment.Currency_COP)

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
