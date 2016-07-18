package models

import (
	"github.com/jinzhu/gorm"
	"math/rand"
	"auth/config"
	"strconv"
	"time"
)

//UserPassword model
type UserPassword struct {
	gorm.Model
	UserID      uint   `gorm:"unique_index"`
	SmsPassword string `gorm:"type:varchar(6)"`
}

//UserPasswords passwords repository
type UserPasswords interface {
	Create(*UserPassword) error
	Delete(*UserPassword) error
	FindByUserID(userID uint) (*UserPassword, error)
}

//UserPasswordsImpl repository implementation
type UserPasswordsImpl struct {
	db *gorm.DB
}

//MakeNewUserPasswords return passwords repository
func MakeNewUserPasswords(db *gorm.DB) UserPasswords {
	return &UserPasswordsImpl{db: db}
}

//Create creates a new model in the db and generates a password
func (up *UserPasswordsImpl) Create(m *UserPassword) error {
	m.SmsPassword = generateRandomPass(config.Get().PasswordLen)
	return up.db.Create(m).Error
}

//Delete deletes the model from db
func (up *UserPasswordsImpl) Delete(m *UserPassword) error {
	return up.db.Unscoped().Delete(m).Error
}

//FindByUserID returns the password by user id
func (up *UserPasswordsImpl) FindByUserID(userID uint) (*UserPassword, error) {
	m := &UserPassword{}
	db := up.db.Where("user_id = ?", userID).First(m)
	if db.RecordNotFound() {
		return nil, nil
	}
	return m, db.Error
}

func generateRandomPass(length int) string {
	rand.Seed(time.Now().UnixNano())
	pass := ""
	for i := 0; i < length; i++ {
		next := rand.Int63n(10)
		pass += strconv.FormatInt(next, 10)
	}
	return pass
}
