package models

import (
	"github.com/jinzhu/gorm"
	"proto/chat"
	"utils/db"
)

//Member is representation of conversation member
type Member struct {
	db.Model
	UserID         uint64 `gorm:"index; unique_index:once_per_conv"`
	Role           string
	ConversationID uint64 `grom:"index; unique_index:once_per_conv"`
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
		Id:            m.ID,
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
			Model: db.Model{
				ID: reqMember.Id,
			},
			UserID:      reqMember.UserId,
			Role:        reqMember.Role.String(),
			Name:        reqMember.Name,
			InstagramID: reqMember.InstagramId,
		})
	}
	return
}

func (m *memberRepository) UpdateLastMessageID(memberID, messageID uint64) error {
	return m.db.Model(&Member{}).Where("id = ?", memberID).Update("last_message_id", messageID).Error
}
