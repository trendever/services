package models

import (
	"fmt"

	"proto/payment"

	"common/db"
	"github.com/durango/go-credit-card"
)

// Payment defines payment order info
type Payment struct {
	db.Model

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

	// redirect URL template. Format string: 1st %v -- success bool; 2nd -- lead id (may be zero)
	Redirect string `gorm:"text"`
}

// GetPayByID returns payment by ID
func (r *RepoImpl) GetPayByID(id uint64) (*Payment, error) {
	var result Payment
	err := r.DB.Where("id = ?", id).Find(&result).Error
	return &result, err
}

// FinishedSessionsForPayID returns num of successfull payments with given pay ID
func (r *RepoImpl) FinishedSessionsForPayID(payID uint64) (int, error) {

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
func (r *RepoImpl) UpdateServiceData(id uint64, data string) error {
	return r.DB.Model(&Payment{}).Where("id = ?", id).Update("service_data", data).Error
}

// CanCreateOrder shows if you can create another order for this leadID
func (r *RepoImpl) CanCreateOrder(leadID uint64) (bool, error) {

	if leadID == 0 {
		return true, nil
	}

	// 1: can create if there are not uncancelled orders
	var ids []uint64
	err := r.DB.
		Model(&Payment{}).
		Where("lead_id = ? and not cancelled", leadID).
		Pluck("id", &ids).
		Error
	if err != nil {
		return false, err
	}
	if len(ids) == 0 {
		return true, nil
	}

	// 2: we have orders; they all are either opened, either successfull. cancelled ones are already filtered
	var (
		sessions []Session
		unclosed = map[uint64]bool{}
	)

	// mark all the leads should be closed if we really want to create a new one
	for _, id := range ids {
		unclosed[id] = true
	}

	err = r.DB.
		Model(&Session{}).
		Where("payment_id in (?)", ids).
		Find(&sessions).
		Error

	if err != nil {
		return false, err
	}

	for _, sess := range sessions {
		if sess.Success {
			delete(unclosed, sess.PaymentID)
		}
	}

	return len(unclosed) == 0, nil
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
		Redirect:       pay.Redirect,

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
