package models

import (
	"github.com/jinzhu/gorm"
	"proto/chat"
	"time"
	"utils/db"
)

//Member is representation of conversation member
type Member struct {
	UserID         uint64 `gorm:"primary_key;index"`
	ConversationID uint64 `gorm:"primary_key;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `sql:"index"`
	Role           string
	Messages       []Message
	Name           string
	LastMessageID  uint64
	// effective id to answers, may be id of shop instead of user itself
	InstagramID uint64
}

//MemberRepository is interface of members' repository
type memberRepository struct {
	db *gorm.DB
}

//Encode converts member to protobuf model
func (m *Member) Encode() *chat.Member {
	role, _ := chat.MemberRole_value[m.Role]
	return &chat.Member{
		UserId:        m.UserID,
		Role:          chat.MemberRole(role),
		Name:          m.Name,
		LastMessageId: m.LastMessageID,
		InstagramId:   m.InstagramID,
	}
}

//DecodeMember converts members of protobuf to model of member
func DecodeMember(pbmembers ...*chat.Member) (members []*Member) {
	for _, reqMember := range pbmembers {
		members = append(members, &Member{
			UserID:      reqMember.UserId,
			Role:        reqMember.Role.String(),
			Name:        reqMember.Name,
			InstagramID: reqMember.InstagramId,
		})
	}
	return
}

func (m *Member) UpdateLastMessageID(messageID uint64) error {
	return db.New().Model(m).Update("last_message_id", messageID).Error
}
