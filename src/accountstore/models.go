package main

import (
	"proto/accountstore"
	"utils/db"
)

// Account contains instagram account cookie
type Account struct {
	InstagramUsername string `gorm:"primary_key"`
	Cookie            string `gorm:"text"`
	Valid             bool   `sql:"default:false"`
}

// AccountRepo generic db access
type AccountRepo interface {
	Create(*Account) error
	Save(*Account) error
	FindValid() ([]Account, error)
	FindByName(string) (*Account, error)
}

// AccountRepoImpl is real db access
type AccountRepoImpl struct {
}

// Create new acc
func (r *AccountRepoImpl) Create(acc *Account) error {
	return db.New().Create(acc).Error
}

// Save it
func (r *AccountRepoImpl) Save(acc *Account) error {
	return db.New().Save(acc).Error
}

// FindValid returns valid only
func (r *AccountRepoImpl) FindValid() ([]Account, error) {
	var out []Account
	err := db.New().Where("valid != FALSE").Find(&out).Error
	return out, err
}

// FindByName returns account by username
func (r *AccountRepoImpl) FindByName(name string) (*Account, error) {
	var out Account
	err := db.New().Where("instagram_username = ?", name).Find(&out).Error
	return &out, err
}

// EncodeAll encodes array to protobuf model
func EncodeAll(accs []Account) []*accountstore.Account {
	out := make([]*accountstore.Account, len(accs))
	for i, acc := range accs {
		out[i] = acc.Encode()
	}
	return out
}

// Encode encodes to protobuf model
func (acc *Account) Encode() *accountstore.Account {
	return &accountstore.Account{
		InstagramUsername: acc.InstagramUsername,
		Cookie:            acc.Cookie,
		Valid:             acc.Valid,
	}
}

// EncodePrivate encodes to protobuf model; hide sensitive fields
func (acc *Account) EncodePrivate() *accountstore.Account {
	return &accountstore.Account{
		InstagramUsername: acc.InstagramUsername,
		Valid:             acc.Valid,
	}
}
