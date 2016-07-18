package models

import (
	"github.com/jinzhu/gorm"
	"strings"
)

//ScopeFunc is a interface for scope functions
type ScopeFunc func(db *gorm.DB) *gorm.DB

//InstagramUsernameScope applies filter by field instagram_username
func InstagramUsernameScope(name string) ScopeFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("instagram_username = ?", strings.ToLower(name))
	}
}
