package models

import (
	"strconv"
	"strings"
	"time"
	"utils/log"
)

type RequestType uint64

const (
	SendMessageRequest RequestType = iota
	ShareMediaRequest
	CreateThreadRequest
)

type DirectRequest struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt time.Time

	Type   RequestType
	UserID uint64
	//@CHECK i'm somewhat unsure whether ReplyKey should be unique
	ReplyKey string // `gorm:"unique" sql:"default:NULL"`

	ThreadID           string
	Participants       []uint64 `gorm:"-"`
	ParticipantsPacked string

	Data    string `gorm:"text"`
	Caption string `gorm:"text"`
}

func (r *DirectRequest) BeforeSave() {
	r.ParticipantsPacked = ""
	for _, p := range r.Participants {
		r.ParticipantsPacked = r.ParticipantsPacked + strconv.FormatUint(p, 10) + ", "
	}
	r.ParticipantsPacked = strings.TrimSuffix(r.ParticipantsPacked, ", ")
}

func (r *DirectRequest) AfterFind() {
	r.Participants = []uint64{}
	strs := strings.Split(r.ParticipantsPacked, ",")
	for _, str := range strs {
		id, err := strconv.ParseUint(strings.Trim(str, " \r\t\n"), 10, 64)
		if err != nil {
			log.Errorf("invalid format in ParticipantsPacked: %v:", str)
			continue
		}
		r.Participants = append(r.Participants, id)
	}
}
