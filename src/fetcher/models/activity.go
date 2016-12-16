package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/bot"
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
	UserID            uint64 // commentary owner ID
	UserName          string // commentary owner username
	UserImageURL      string
	MentionedUsername string // mention tag. @saveit, for instance
	MentionedRole     bot.MentionedRole
	Type              string
	Comment           string
	ThreadID          string `gorm:"index"` // optional field: direct thread ID
}

// TableName fixes this model table name
func (act *Activity) TableName() string {
	return "activities_activity"
}

// Create activity to db if new
func (act *Activity) Create() error {
	// write activity to DB
	var count int

	// check by pk if record exist
	err := db.New().Model(&Activity{}).Where("pk = ?", act.Pk).Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check whether activity already exists: %v", err)
	}

	if count > 0 {
		// skipping dupe
		return nil
	}

	// now -- create
	err = db.New().Create(act).Error
	if err != nil {
		return fmt.Errorf("failed to save activity: %v", err)
	}

	log.Debug("Add row: %v", act.Pk)
	return nil
}

// Save existing activity
func (act *Activity) Save() error {
	return db.New().Save(act).Error
}

// Encode to protobuf
func (act *Activity) Encode() *bot.Activity {
	return &bot.Activity{
		Id:                int64(act.ID),
		Pk:                act.Pk,
		MediaId:           act.MediaID,
		MediaUrl:          act.MediaURL,
		UserId:            act.UserID,
		UserImageUrl:      act.UserImageURL,
		UserName:          act.UserName,
		MentionedUsername: act.MentionedUsername,
		MetionedRole:      act.MentionedRole,
		Type:              act.Type,
		Comment:           act.Comment,
		CreatedAt:         act.CreatedAt.Unix(),
		DirectThreadId:    act.ThreadID,
	}
}

//EncodeActivities encodes activity arr to protobuf
func EncodeActivities(activities []Activity) []*bot.Activity {

	out := make([]*bot.Activity, len(activities))

	for i := range activities {
		out[i] = (&activities[i]).Encode()
	}

	return out
}
