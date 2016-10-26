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

	api, err := instagram.NewInstagram(login, password)
	if err != nil {
		return nil, err
	}

	cookieJar, err := api.Save()
	if err != nil {
		return nil, err
	}

	return &Account{
		InstagramUsername: login,
		Cookie:            cookieJar,
		Valid:             true, // @TODO ofc
	}, nil
}

// VerifyCode is verification process; can fail -- no err returned, but given account is still marked as invalid
func (r *InstagramAccessImpl) VerifyCode(*Account) error {

	return fmt.Errorf("Error! Not implemented")
}
