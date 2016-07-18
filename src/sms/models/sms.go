package models

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
)

//SmsJSON is json model
type SmsJSON struct {
	Error  string          `json:"error"`
	Code   string          `json:"code"`
	Result json.RawMessage `json:"result"`
}

//ResultSuccess is json model of success result
type ResultSuccess struct {
	ID       int64   `json:"id"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

//SmsDB is database model
type SmsDB struct {
	gorm.Model
	Message   string `gorm:"type:text"`
	Phone     string
	SmsID     int64  // id from atompark
	SmsStatus string // our status
	SmsError  string `gorm:"type:text"` // error from atompark if exist
}

//TableName sets table name
func (s *SmsDB) TableName() string {
	return "sms"
}

//SmsRepository is interface for working with SmsDB models
type SmsRepository interface {
	Create(*SmsDB) error
	Update(*SmsDB) error
	GetByID(uint) (*SmsDB, error)
}

//SmsRepositoryImpl implements SmsRepository
type SmsRepositoryImpl struct {
	db *gorm.DB
}

//MakeNewSmsRepository returns new SmsRepository
func MakeNewSmsRepository(db *gorm.DB) SmsRepository {
	return &SmsRepositoryImpl{db: db}
}

//Create creates record in the DB
func (s *SmsRepositoryImpl) Create(sms *SmsDB) error {
	return s.db.Create(sms).Error
}

//Update updates record in the DB
func (s *SmsRepositoryImpl) Update(sms *SmsDB) error {
	return s.db.Save(sms).Error
}

//GetByID returns SmsDB model by it's ID
func (s *SmsRepositoryImpl) GetByID(id uint) (*SmsDB, error) {
	sms := &SmsDB{}
	db := s.db.First(sms, id)
	if db.RecordNotFound() {
		return nil, nil
	}
	return sms, db.Error
}
