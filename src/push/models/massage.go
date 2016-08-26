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
	Data      string `gorm:"text"`
	Body      string `gorm:"text"`
	Title     string `gorm:"text"`
	Priority  push.Priority
	Receivers []Receiver `gorm:"many2many:messages_receivers"`
}

type Receiver struct {
	ID      uint64 `gorm:"primary_key"`
	Service push.ServiceType
	Token   string
}
