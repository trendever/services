package models

import (
	"fmt"
	"utils/db"
)

// ThreadInfo tells that this thread was processed
type ThreadInfo struct {
	ThreadID      string `gorm:"primary_key"`
	LastCheckedID string
	Notified      bool
}

// LaterThan tells if info.LastCheckedID mark is placed later than otherID (so, info.LastCheckedID should have ID that is less than otherID)
func (info *ThreadInfo) LaterThan(otherID string) bool {
	// info  other result
	// ""    "5"   false
	// "100" "0"   true
	// "150" "100" true
	lenDiff := len(info.LastCheckedID) - len(otherID)
	if lenDiff != 0 {
		return lenDiff < 0
	}
	return info.LastCheckedID < otherID
}

// TableName defines table name
func (info *ThreadInfo) TableName() string {
	return "direct_thread"
}

// GetThreadInfo gives us thread info for this threadID
func GetThreadInfo(threadID string) (ThreadInfo, error) {

	var res ThreadInfo
	err := db.New().
		FirstOrCreate(&res, ThreadInfo{ThreadID: threadID}).
		Error

	return res, err
}

// SaveLastCheckedID updates threadInfo with lastCheckedID=messageID
func SaveLastCheckedID(threadID, messageID string) error {
	err := db.New().
		Model(&ThreadInfo{}).
		Where("thread_id = ?", threadID).
		Update("last_checked_id", messageID).
		Error
	if err != nil {
		return fmt.Errorf("failed to save last checked id for thread %v: %v", threadID, err)
	}
	return nil
}

// Save just saves it in usual way
func (info *ThreadInfo) Save() error {
	return db.New().Save(info).Error
}
