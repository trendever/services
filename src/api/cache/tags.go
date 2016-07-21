package cache

import "time"

type Tags map[string]string

func AddTags(key string, ttl time.Duration, tags ...string) {
	tt := make(Tags)
	GetV(key, &tt)
	tm := time.Now()
	for _, t := range tags {
		tt[t] = tm.String()
	}
	Put(key, tt, ttl)
}

func GetTags(key string) []string {
	tags := make(Tags)
	GetV(key, &tags)
	out := make([]string, len(tags), len(tags))
	i := 0
	for k := range tags {
		out[i] = k
		i++
	}
	return out
}
