package main

import (
	"proto/accountstore"
	"utils/db"
)

// Account contains instagram account cookie
type Account struct {
	InstagramUsername string `gorm:"primary_key"`
	Role              accountstore.Role
	Cookie            string `gorm:"text"`
	Valid             bool   `sql:"default:false"`
	CodeSent          int64
	CodeSentBy        string
}

// Create new acc
func Create(acc *Account) error {
	return db.New().Create(acc).Error
}

// Save it
func Save(acc *Account) error {
	return db.New().Save(acc).Error
}

// Find returns valid only
func Find(validOnly bool, roles []accountstore.Role) ([]Account, error) {
	var out []Account
	req := db.New()

	if validOnly {
		req = req.Where("valid != FALSE")
	}

	if len(roles) > 0 {
		req = req.Where("role in (?)", roles)
	}

	err := req.Find(&out).Error
	return out, err
}

// FindByName returns account by username
func FindByName(name string) (*Account, error) {
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
		Role:              acc.Role,
	}
}

// EncodePrivate encodes to protobuf model; hide sensitive fields
func (acc *Account) EncodePrivate() *accountstore.Account {
	return &accountstore.Account{
		InstagramUsername: acc.InstagramUsername,
		Valid:             acc.Valid,
		Role:              acc.Role,
	}
}
