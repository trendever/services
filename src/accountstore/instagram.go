package main

import (
	"fmt"
	"instagram"
)

// InstagramAccess is mockable instagram adapter
type InstagramAccess interface {
	Login(login, password string) (*Account, error)
	VerifyCode(*Account) error
}

// InstagramAccessImpl is real instagram connector
type InstagramAccessImpl struct {
}

// Login with given login:pass, return an Account (probably invalid -- confirmation needed)
func (r *InstagramAccessImpl) Login(login, password string) (*Account, error) {

	account := &Account{
		InstagramUsername: login,
		Valid:             true,
	}

	api, err := instagram.NewInstagram(login, password)
	if err == instagram.ErrorCheckpointRequired {
		account.Valid = false
	} else if err != nil {
		return nil, err
	}

	cookieJar, err := api.Save()
	if err != nil {
		return nil, err
	}

	account.Cookie = cookieJar

	return account, nil
}

// VerifyCode is verification process; can fail -- no err returned, but given account is still marked as invalid
func (r *InstagramAccessImpl) VerifyCode(*Account) error {

	return fmt.Errorf("Error! Not implemented")
}
