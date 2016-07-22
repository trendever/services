package cache

import (
	"api/conf"
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"time"
	"utils/log"
)

var c *redis.Client

//Init initializes connection to the cache
func Init() {
	conf := conf.GetSettings()
	if conf.Redis.Addr == "" {
		return
	}
	cc := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	_, err := cc.Ping().Result()

	log.Fatal(err)

	c = cc
}

//Put puts the value to the key in the cache
func Put(key string, value interface{}, ttl time.Duration) error {
	if c == nil {
		return nil
	}
	d, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(key, d, ttl).Err()
}

//Get gets a value by the key
func Get(key string) (interface{}, error) {
	if c == nil {
		return nil, nil
	}
	val, err := c.Get(key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func GetV(key string, v interface{}) error {
	ret, err := Get(key)
	if err != nil {
		return err
	}

	if ret == nil {
		return nil
	}

	return json.Unmarshal([]byte(ret.(string)), v)
}

//Flush deletes all data from db
func Flush() error {
	if c == nil {
		return fmt.Errorf("Connection not configured")
	}
	return c.FlushDb().Err()
}

//Delete deletes the keys from the cache
func Delete(keys ...string) error {
	if c == nil {
		return nil
	}

	return c.Del(keys...).Err()
}
