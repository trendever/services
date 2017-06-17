package models

import (
	"utils/db"
)

// ThreadInfo tells that this thread was processed
type ThreadInfo struct {
	ThreadID      string `gorm:"primary_key"`
	SourceID      uint64 `gorm:"primary_key"`
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

func CompareID(one, two string) int {
	lenDiff := len(one) - len(two)
	if lenDiff != 0 {
		if lenDiff > 0 {
			return 1
		}
		return -1
	}
	if one == two {
		return 0
	}
	if one > two {
		return 1
	}
	return -1
}

// TableName defines table name
func (info *ThreadInfo) TableName() string {
	return "direct_thread"
}

// GetThreadInfo gives us thread info for this threadID
func GetThreadInfo(threadID string, sourceID uint64) (ThreadInfo, error) {

	var res ThreadInfo
	err := db.New().
		FirstOrInit(&res, ThreadInfo{ThreadID: threadID, SourceID: sourceID}).
		Error

	return res, err
}

// Save just saves it in usual way
func (info *ThreadInfo) Save() error {
	return db.New().Save(info).Error
}
