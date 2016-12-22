package main

import (
	"time"

	"proto/accountstore"
	"sync"
	"utils/db"
	"utils/log"
	"utils/nats"
)

const notifyTopic = "accountstore.notify"

var global struct {
	once       sync.Once
	notifyChan chan *Account
}

// Account contains instagram account cookie
type Account struct {
	InstagramUsername string `gorm:"primary_key"`
	InstagramID       uint64 `gorm:"index"`
	CreatedAt         time.Time
	Role              accountstore.Role
	Cookie            string `gorm:"text"`
	Valid             bool   `sql:"default:false"`
	OwnerID           uint64 `gorm:"index"`
	CodeSent          int64
	CodeSentBy        string
}

// Save it
func Save(acc *Account) error {
	err := db.New().Save(acc).Error
	if err != nil {
		return err
	}
	global.notifyChan <- acc
	return nil
}

func notifier() {
	global.once.Do(func() {
		global.notifyChan = make(chan *Account, 20)
		// @TODO notify about service (re-)start? resend all accounts or just tell clients to reload all them
		for acc := range global.notifyChan {
			for {
				err := nats.StanPublish(notifyTopic, acc)
				if err != nil {
					log.Errorf("failed to notify about account: %v", err)
					time.Sleep(time.Second * 5)
				} else {
					break
				}
			}
		}
	})
}

// Find returns valid only
func Find(in *accountstore.SearchRequest) ([]Account, error) {
	var out []Account
	req := db.New()

	if in.InstagramUsername > "" {
		req = req.Where("instagram_username = ?", in.InstagramUsername)
	}

	if in.InstagramId > 0 {
		req = req.Where("instagram_id = ?", in.InstagramId)
	}

	if in.OwnerId > 0 {
		req = req.Where("owner_id = ?", in.OwnerId)
	}

	if !in.IncludeInvalids {
		req = req.Where("valid != FALSE")
	}

	if len(in.Roles) > 0 {
		req = req.Where("role in (?)", in.Roles)
	}

	err := req.Find(&out).Error
	return out, err
}

// FindAccount returns account by template
func FindAccount(template *Account) (*Account, error) {
	var out Account
	err := db.New().Where(template).Find(&out).Error
	return &out, err
}

// EncodeAll encodes array to protobuf model
func EncodeAll(accs []Account, private bool) []*accountstore.Account {
	out := make([]*accountstore.Account, len(accs))
	for i, acc := range accs {
		if private {
			out[i] = acc.EncodePrivate()
		} else {
			out[i] = acc.Encode()
		}
	}
	return out
}

// Encode encodes to protobuf model
func (acc *Account) Encode() *accountstore.Account {
	return &accountstore.Account{
		InstagramUsername: acc.InstagramUsername,
		InstagramId:       acc.InstagramID,
		OwnerId:           acc.OwnerID,
		Cookie:            acc.Cookie,
		Valid:             acc.Valid,
		Role:              acc.Role,
		CreatedAt:         acc.CreatedAt.Unix(),
		CreatedAtAgo:      uint64(time.Since(acc.CreatedAt).Seconds()),
	}
}

// EncodePrivate encodes to protobuf model; hide sensitive fields
func (acc *Account) EncodePrivate() *accountstore.Account {
	return &accountstore.Account{
		InstagramUsername: acc.InstagramUsername,
		InstagramId:       acc.InstagramID,
		Valid:             acc.Valid,
		Role:              acc.Role,
		CreatedAt:         acc.CreatedAt.Unix(),
		CreatedAtAgo:      uint64(time.Since(acc.CreatedAt).Seconds()),
	}
}
