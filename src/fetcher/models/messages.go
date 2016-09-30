package models

import (
	"github.com/jinzhu/gorm"
)

// Thread defines instagram direct thread
// model used to keep track of processed messages
type Thread struct {
	gorm.Model

	// instagram's post primary key from json
	ThreadID      string `gorm:"index"`
	LastMessageID string
	Replied       bool
}

// TableName defines table name
func (t *Thread) TableName() string {
	return "direct_threads"
}
