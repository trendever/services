package models

import (
	"proto/push"
	"time"
)

type Message struct {
	ID         uint64    `gorm:"primary_key"`
	Expiration time.Time `gorm:"index"`
	// last send try time
	LastTry   time.Time
	Body      string
	Priority  push.Priority
	Receivers []Receiver `gorm:"many2many:messages_receivers"`
}

// this func do NOT copy receivers from source
func DecodeMessage(in *push.PushMessage) *Message {
	return &Message{
		Body:       in.Body,
		Priority:   in.Prority,
		Expiration: time.Now().Add(time.Second * time.Duration(in.TimeToLive)),
	}
}

type Receiver struct {
	ID      uint64 `gorm:"primary_key"`
	Service push.ServiceType
	Token   string
}
