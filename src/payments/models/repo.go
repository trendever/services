package models

import (
	"github.com/jinzhu/gorm"
)

// Repo is mockable payment repository
type Repo interface {

	// pay part
	GetPayByID(uint) (*Payment, error)
	CreatePay(*Payment) error
	SavePay(*Payment) error
	CanCreateOrder(leadID uint) (bool, error)
	UpdateServiceData(uint, string) error

	// sess part
	CreateSess(*Session) error
	GetSessByUID(string) (*Session, error)
	FinishedSessionsForPayID(pay uint) (int, error)
	SaveSess(*Session) error
	GetUnfinished() ([]Session, error)
}

// RepoImpl is implementation that uses gorm
type RepoImpl struct {
	DB *gorm.DB
}
