package models

import (
	"proto/push"
	"time"
)

type PushNotify struct {
	ID         uint64    `gorm:"primary_key"`
	Expiration time.Time `gorm:"index"`
	// last send try time
	LastTry   time.Time
	Data      string `gorm:"text"`
	Body      string `gorm:"text"`
	Title     string `gorm:"text"`
	Priority  push.Priority
	Receivers []PushReceiver `gorm:"many2many:push_notifies_receivers"`
}

type PushReceiver struct {
	ID      uint64 `gorm:"primary_key"`
	Service push.ServiceType
	Token   string
}

// returns service -> []tokens map of receivers
func (n *PushNotify) MapReceivers() map[push.ServiceType][]string {
	tokens := make(map[push.ServiceType][]string)
	for _, receiver := range n.Receivers {
		tokens[receiver.Service] = append(tokens[receiver.Service], receiver.Token)
	}
	return tokens
}

func (n *PushNotify) ReceiversFromMap(receivers map[push.ServiceType][]string) {
	n.Receivers = nil
	for service, tokens := range receivers {
		for _, token := range tokens {
			n.Receivers = append(n.Receivers, PushReceiver{
				Service: service,
				Token:   token,
			})
		}
	}
}

func DecodeNotify(in *push.PushRequest) *PushNotify {
	notify := &PushNotify{
		Expiration: time.Now().Add(time.Second * time.Duration(in.Message.TimeToLive)),
		Data:       in.Message.Data,
		Body:       in.Message.Body,
		Title:      in.Message.Title,
		Priority:   in.Message.Priority,
	}
	for _, receiver := range in.Receivers {
		notify.Receivers = append(notify.Receivers, PushReceiver{
			Service: receiver.Service,
			Token:   receiver.Token,
		})
	}
	return notify
}
