package models

import (
	"common/db"
	"proto/payment"
)

// Session once-used pay sess
type Session struct {
	db.Model

	PaymentID uint64
	Payment   *Payment

	Amount      uint64
	IP          string
	GatewayType string `gorm:"index"`

	State         string `gorm:"index"`
	Finished      bool   `gorm:"index" sql:"default:false"` // session can be finished, but unsuccessfully
	Success       bool   `gorm:"index" sql:"default:false"`
	FailureReason string

	// I wonder why payture wants 2 unique ids;
	UniqueID   string `gorm:"index"` // this one is used to check pay status
	ExternalID string `gorm:"index"` // this one is used by client
}

// CreateSess to DB
func (r *RepoImpl) CreateSess(s *Session) error {
	return r.DB.Create(s).Error
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

// GetUnfinished returns payment by ID
func (r *RepoImpl) GetUnfinished() ([]Session, error) {
	var result []Session

	err := r.DB.
		Where("finished != TRUE").
		Preload("Payment").
		Find(&result).
		Error

	return result, err
}

// SaveSess saves payment
func (r *RepoImpl) SaveSess(p *Session) error {
	return r.DB.Save(p).Error
}

// SavePay saves payment
func (r *RepoImpl) SavePay(p *Payment) error {
	return r.DB.Save(p).Error
}

// Info get user info for this order
func (sess *Session) Info() *payment.UserInfo {

	var info = &payment.UserInfo{}
	if sess.Payment != nil {
		info = sess.Payment.Info()
	}

	if sess.IP != "" {
		info.Ip = sess.IP
	}

	return info
}
