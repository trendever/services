package models

import (
	"github.com/jinzhu/gorm"
)

type Activity struct {
	gorm.Model

	// instagram's post primary key from json
	Pk                string `gorm:"not null;unique"`
	MediaID           string
	MediaUrl          string
	UserID            int64  // commentary owner ID
	UserName          string // commentary owner username
	UserImageUrl      string
	MentionedUsername string // mention tag. @saveit, for instance
	Type              string
	Comment           string
}

// set table name
func (this *Activity) TableName() string {
	return "activities_activity"
}
