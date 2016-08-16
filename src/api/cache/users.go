package cache

import (
	"utils/log"
)

// FlushUser flushes user cache
func FlushUser(id int64) {
	log.Debug("Flushing user %v", id)
	flush(idKey("user", id))
}

// FlushShop flushes shop cache
func FlushShop(id int64) {
	log.Debug("Flushing shop %v", id)
	flush(idKey("shop", id))
}
