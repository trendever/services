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

	UserID    uint64 // that's client id
	LeadID    uint64 `gorm:"index"`
	Cancelled bool   `sql:"default:false"` // sender side can cancel payment

	GatewayType string // implemented gate, like `payture`
	ServiceName string `gorm:"index"` // like `api` or `trendcoin`
	ServiceData string `gorm:"text"`
	Comment     string `gorm:"text"`

	// p2p params
	ShopCardNumber string
	Amount         uint64
	Currency       int32

	// non-p2p params
	CardID string

	// in coins
	CommissionFee uint64
	// user id, usually supplier
	CommissionSource uint64
}

// GetPayByID returns payment by ID
func (r *RepoImpl) GetPayByID(id uint) (*Payment, error) {
	var result Payment
	err := r.DB.Where("id = ?", id).Find(&result).Error
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

// CreatePay creates payment
func (r *RepoImpl) CreatePay(p *Payment) error {
	return r.DB.Create(p).Error
}

// UpdateServiceData service data can be upgraded
func (r *RepoImpl) UpdateServiceData(id uint, data string) error {
	return r.DB.Model(&Payment{}).Where("id = ?", id).Update("service_data", data).Error
}

// CanCreateOrder shows if you can create another order for this leadID
func (r *RepoImpl) CanCreateOrder(leadID uint) (bool, error) {

	if leadID == 0 {
		return true, nil
	}

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
	info := r.Info

	// Check credit cards first
	if data.ShopCardNumber != "" && !validate(data.ShopCardNumber) {
		return nil, fmt.Errorf("Invalid credit card (%v) (Luhn failed)", data.ShopCardNumber)
	}

	if data.CommissionFee != 0 && data.CommissionSource == 0 {
		return nil, fmt.Errorf("Empty commission source")
	}

	if data.Gateway == "" {
		return nil, fmt.Errorf("Empty gateway info")
	}

	pay := DecodePayment(data)

	if info != nil {
		pay.UserID = info.UserId
	}

	return pay, nil
}

// DecodePayment protobuf.OrderData -> models.Payment
func DecodePayment(pay *payment.OrderData) *Payment {
	return &Payment{
		// Bank cards
		ShopCardNumber: pay.ShopCardNumber,
		Cancelled:      pay.Cancelled,
		GatewayType:    pay.Gateway,
		ServiceName:    pay.ServiceName,
		ServiceData:    pay.ServiceData,
		Comment:        pay.Comment,
		CardID:         pay.CardId,

		// Money
		Amount:   pay.Amount,
		Currency: int32(pay.Currency),
		LeadID:   pay.LeadId,

		CommissionFee:    pay.CommissionFee,
		CommissionSource: pay.CommissionSource,
	}
}

// Encode models.Payment -> protobuf.OrderData
func (pay *Payment) Encode() *payment.OrderData {
	return &payment.OrderData{
		// Bank cards
		ShopCardNumber: pay.ShopCardNumber,
		Cancelled:      pay.Cancelled,
		Gateway:        pay.GatewayType,
		ServiceName:    pay.ServiceName,
		ServiceData:    pay.ServiceData,
		Comment:        pay.Comment,
		CardId:         pay.CardID,

		// Money
		Amount:   pay.Amount,
		Currency: payment.Currency(pay.Currency),
		LeadId:   pay.LeadID,

		CommissionFee:    pay.CommissionFee,
		CommissionSource: pay.CommissionSource,
	}
}

// Info get user info for this order
func (pay *Payment) Info() *payment.UserInfo {
	return &payment.UserInfo{
		UserId: pay.UserID,
	}
}

// decorator to call CC check
func validate(cc string) bool {
	card := creditcard.Card{Number: cc}
	return card.ValidateNumber()
}
