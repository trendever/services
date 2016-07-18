package models

import (
	"core/db"
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
