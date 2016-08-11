package cache

import (
	"fmt"
	"time"
)

const keyPrefix = "tags#"

// AddTags add tags for this key
func AddTags(key string, ttl time.Duration, tags ...string) {
	for _, t := range tags {
		c.HSet(keyPrefix+t, key, "")
	}

	c.Expire(key, ttl)
}

// GetTags returns previously added keys
func GetTags(key string) []string {
	res, err := c.HKeys(keyPrefix + key).Result()
	if err != nil {
		return nil
	}

	return res
}

func idKey(name string, id int64) string {
	return fmt.Sprintf("%v.%v", name, id)
}
