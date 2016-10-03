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

// GreaterOrEqual returns true if this thread info has ID greater than some other (given)
func (info *ThreadInfo) GreaterOrEqual(otherID string) bool {
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
