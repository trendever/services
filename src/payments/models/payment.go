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

	// in coins
	CommissionFee uint64
	// user id, usually supplier
	CommissionSource uint64
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
		Where("finished = true").
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
		Joins("LEFT JOIN sessions as sess ON payments.id = sess.payment_id").
		Where("lead_id = ? and not cancelled and (finished ISNULL or not finished)", leadID).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// NewPayment generates a payment infocard by request
func NewPayment(r *payment.CreateOrderRequest) (*Payment, error) {
	data := r.Data
	// Check credit cards first
	if !validate(data.ShopCardNumber) {
		return nil, fmt.Errorf("Invalid credit card (Luhn failed)")
	}

	if data.LeadId <= 0 {
		return nil, fmt.Errorf("Empty lead ID")
	}

	if data.CommissionFee != 0 && data.CommissionSource == 0 {
		return nil, fmt.Errorf("Empty commission source")
	}

	pay := &Payment{
		// Bank cards
		ShopCardNumber: data.ShopCardNumber,

		// Core LeadID
		LeadID:         data.LeadId,
		Direction:      int32(data.Direction),
		ConversationID: data.ConversationId,
		UserID:         data.UserId,

		CommissionFee:    data.CommissionFee,
		CommissionSource: data.CommissionSource,
	}

	switch data.Currency {
	case payment.Currency_RUB:

		// must convert to cops (1/100 of rub)
		pay.Amount = data.Amount * 100
		pay.Currency = int32(payment.Currency_COP)

	case payment.Currency_COP:

		pay.Amount = data.Amount
		pay.Currency = int32(payment.Currency_COP)

	default:
		// unknown currency! panic
		return nil, fmt.Errorf("Unsupported currency %v (%v)", data.Currency, payment.Currency_name[int32(data.Currency)])
	}

	return pay, nil
}

// decorator to call CC check
func validate(cc string) bool {
	card := creditcard.Card{Number: cc}
	return card.ValidateNumber()
}
