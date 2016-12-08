package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"instagram"
	"time"
)

// InstagramAccess is mockable instagram adapter
type InstagramAccess interface {
	Login(login, password string, preferEmail bool) (*Account, error)
	SendCode(*Account, string, bool) error
	VerifyCode(*Account, string) error
}

// InstagramAccessImpl is real instagram connector
type InstagramAccessImpl struct {
}

// Login with given login:pass, return an Account (probably invalid -- confirmation needed)
func (r *InstagramAccessImpl) Login(login, password string, preferEmail bool) (*Account, error) {

	var account *Account

	// find existing or create new
	if found, err := FindAccount(&Account{InstagramUsername: login}); err == gorm.ErrRecordNotFound {
		account = &Account{
			InstagramUsername: login,
			Valid:             true,
		}
	} else if err != nil {
		return nil, err
	} else {
		account = found
	}

	var (
		api *instagram.Instagram
		err error
	)

	if account.Cookie > "" {
		api, err = instagram.Restore(account.Cookie, password)
	} else {
		api, err = instagram.NewInstagram(login, password)
	}

	if err == instagram.ErrorCheckpointRequired {
		account.Valid = false

		err := r.sendCode(api, account, preferEmail)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	cookieJar, err := api.Save()
	if err != nil {
		return nil, err
	}

	account.Cookie = cookieJar
	account.InstagramID = api.UserNameID
	account.Valid = true

	return account, nil
}

func (r *InstagramAccessImpl) sendCode(api *instagram.Instagram, acc *Account, preferEmail bool) error {
	sentCode, err := api.SendCode(preferEmail)
	if err != nil {
		return err
	}

	acc.CodeSent = time.Now().Unix()
	acc.CodeSentBy = sentCode
	return nil
}

// SendCode sends instagram checkpoint code
func (r *InstagramAccessImpl) SendCode(acc *Account, password string, preferEmail bool) error {

	api, err := instagram.Restore(acc.Cookie, "")
	if err != nil {
		return err
	}

	api.SetPassword(password)
	return r.sendCode(api, acc, preferEmail)
}

// VerifyCode is verification process; can fail -- no err returned, but given account is still marked as invalid
func (r *InstagramAccessImpl) VerifyCode(acc *Account, code string) error {

	api, err := instagram.Restore(acc.Cookie, "")
	if err != nil {
		return err
	}

	if time.Now().Unix()-acc.CodeSent > int64((time.Minute * 15).Seconds()) {
		return fmt.Errorf("Timeout error")
	}

	return api.CheckCode(code)
}
