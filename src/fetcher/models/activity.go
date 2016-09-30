package models

import (
	"github.com/jinzhu/gorm"
	"utils/db"
	"utils/log"
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
func (act *Activity) TableName() string {
	return "activities_activity"
}

// Save activity to db if new
func (act *Activity) Save() error {
	// write activity to DB
	var count int

	// check by pk if record exist
	err := db.New().Model(&Activity{}).Where("pk = ?", act.Pk).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		// skipping dupe
		log.Debug("Skipping dupe (got %v times)", count)
		return nil
	}

	// now -- create
	err = db.New().Create(act).Error
	if err != nil {
		return err
	}

	log.Debug("Add row: %v", act.Pk)
	return nil
}
