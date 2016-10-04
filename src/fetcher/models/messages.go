package models

import (
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

	// info		other		result
	// ""			"5"			false
	// "100"  "0"			true
	// "150"	"100"		true

	diff := len(info.LastCheckedID) - len(otherID)

	switch {
	case diff < 0: // len(thisID) is less
		return false
	case diff > 0: // len(otherID) is less
		return true
	default: // lens are equal; compare byte-by-byte
		return info.LastCheckedID >= otherID
	}

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
	// save LastCheckedID
	return db.New().
		Model(&ThreadInfo{}).
		Where("thread_id = ?", threadID).
		Update("last_checked_id", messageID).
		Error
}
