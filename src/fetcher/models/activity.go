package models

import (
	"github.com/jinzhu/gorm"
)

// Activity is main fetcher element
type Activity struct {
	gorm.Model

	// instagram's post primary key from json
	Pk                string `gorm:"not null;unique"`
	MediaID           string
	MediaURL          string
	UserID            int64  // commentary owner ID
	UserName          string // commentary owner username
	UserImageURL      string
	MentionedUsername string // mention tag. @saveit, for instance
	Type              string
	Comment           string
	ThreadID          string `gorm:"index"` // optional field: direct thread ID
}

// TableName fixes this model table name
func (a *Activity) TableName() string {
	return "activities_activity"
}
