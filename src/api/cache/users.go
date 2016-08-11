package cache

import (
	"fmt"
)

// FlushUser flushes username cache
func FlushUser(instagramname string) {
	key := getUsernameTagKey(instagramname)
	keys := GetTags(key)
	if len(keys) > 0 {
		Delete(keys...)
	}
	Delete(key)
}

func getUsernameTagKey(username string) string {
	return fmt.Sprintf("instagramname.%v", username)
}
