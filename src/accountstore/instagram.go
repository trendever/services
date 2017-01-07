package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"instagram"
	"time"
)

// InstagramAccess is mockable instagram adapter
type InstagramAccess interface {
	Login(login, password string, preferEmail bool, owner uint64) (*Account, error)
	SendCode(*Account, string, bool) error
	VerifyCode(acc *Account, password, code string) error
}

// InstagramAccessImpl is real instagram connector
type InstagramAccessImpl struct {
}

// Login with given login:pass, return an Account (probably invalid -- confirmation needed)
func (r *InstagramAccessImpl) Login(login, password string, preferEmail bool, owner uint64) (*Account, error) {

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
		account.Valid = true
	}

	var (
		api *instagram.Instagram
		err error
	)

	if account.Cookie > "" && owner == account.OwnerID {
		api, err = instagram.Restore(account.Cookie, password, true)
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

	api, err := instagram.Restore(acc.Cookie, "", false)
	if err != nil {
		return err
	}

	api.SetPassword(password)
	err = r.sendCode(api, acc, preferEmail)
	if err != nil {
		return err
	}

	cookieJar, err := api.Save()
	if err != nil {
		return err
	}

	acc.Cookie = cookieJar

	return Save(acc)
}

// VerifyCode is verification process; can fail -- no err returned, but given account is still marked as invalid
func (r *InstagramAccessImpl) VerifyCode(acc *Account, password, code string) error {

	api, err := instagram.Restore(acc.Cookie, password, false)
	if err != nil {
		return err
	}

	// check if already confirmed
	_, err = api.GetRecentActivity()
	if err == instagram.ErrorCheckpointRequired {
		if time.Now().Unix()-acc.CodeSent > int64((time.Minute * 15).Seconds()) {
			return fmt.Errorf("Timeout error")
		}

		err = api.CheckCode(code)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	api.CheckpointURL = ""
	api.CheckpointCookies = nil

	_, err = api.GetRecentActivity()
	if err != nil {
		return err
	}

	cookieJar, err := api.Save()
	if err != nil {
		return err
	}

	acc.Valid = true
	acc.Cookie = cookieJar

	return Save(acc)
}
