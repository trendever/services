package models

import (
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
)

type Mail struct {
	gorm.Model
	Mid     string `gorm:"index"`
	From    string `valid:"required,email_with_name"`
	Subject string `valid:"required"`
	Message string `gorm:"type:text" valid:"required"`
	//To contains one or more email addresses, separated by coma
	To        string `gorm:"type:text" valid:"required,emails"`
	Status    string
	StatusMsg string
}

type MailRepository interface {
	Create(*Mail) error
	Update(*Mail) error
	GetByID(id uint) (*Mail, error)
}

type MailRepositoryImpl struct {
	db *gorm.DB
}

func MakeNewMailRepository(db *gorm.DB) MailRepository {
	return &MailRepositoryImpl{db: db}
}

func (r *MailRepositoryImpl) Create(m *Mail) error {
	if ok, err := m.Validate(); !ok {
		return err
	}
	return r.db.Create(m).Error
}
func (r *MailRepositoryImpl) Update(m *Mail) error {
	return r.db.Save(m).Error
}
func (r *MailRepositoryImpl) GetByID(id uint) (*Mail, error) {
	m := &Mail{}
	db := r.db.First(m, id)
	if db.RecordNotFound() {
		return nil, nil
	}
	return m, db.Error
}

func (m *Mail) Validate() (bool, error) {
	return govalidator.ValidateStruct(m)
}
