package models

import (
	"errors"
	"strings"
	"utils/db"
)

//GetUserByID returns user by ID
func GetUserByID(id uint) (*User, error) {
	user := &User{}
	scope := db.New().Find(user, id)
	return user, scope.Error
}

//GetUserByInstagramName returns User by instagram username
func GetUserByInstagramName(name string) (*User, error) {
	user := &User{}
	err := db.New().Scopes(InstagramUsernameScope(name)).Find(user).Error
	return user, err
}

//FindUserIDByInstagramName returns only user's ID by user's instagram username
func FindUserIDByInstagramName(name string) (uint, error) {
	user, err := GetUserByInstagramName(name)
	return user.ID, err
}

// returns first user which match any of arguments
func FindUserMatchAny(ID, instagramID uint64, name, instagramName, email, phone string) (user *User, found bool, err error) {
	scope := db.New()
	ok := false
	if ID != 0 {
		scope = scope.Where("id = ?", ID)
		ok = true
	} else {
		if instagramID != 0 {
			scope = scope.Or("instagram_id = ?", instagramID)
			ok = true
		}
		if name != "" {
			scope = scope.Or("name = ?", name)
			ok = true
		}
		if instagramName != "" {
			scope = scope.Or("instagram_username = ?", strings.ToLower(instagramName))
			ok = true
		}
		if email != "" {
			scope = scope.Or("email = ?", email)
			ok = true
		}
		if phone != "" {
			scope = scope.Or("phone = ?", phone)
			ok = true
		}
		if !ok {
			return nil, false, errors.New("Empty conditions")
		}
	}
	user = &User{}
	res := scope.Find(user)
	if res.RecordNotFound() {
		return nil, false, nil
	}
	return user, true, nil
}
