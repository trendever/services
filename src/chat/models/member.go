package models

import (
	"github.com/jinzhu/gorm"
	"proto/chat"
)

//Member is representation of conversation member
type Member struct {
	gorm.Model
	UserID         uint `gorm:"index"`
	Role           string
	ConversationID uint `grom:"index"`
	Messages       []Message
	Name           string
	LastMessageID  uint
}

//MemberRepository is interface of members' repository
type memberRepository struct {
	db *gorm.DB
}

//Encode converts member to protobuf model
func (m *Member) Encode() *chat.Member {
	role, _ := chat.MemberRole_value[m.Role]
	return &chat.Member{
		Id:            uint64(m.ID),
		UserId:        uint64(m.UserID),
		Role:          chat.MemberRole(role),
		Name:          m.Name,
		LastMessageId: uint64(m.LastMessageID),
	}
}

//DecodeMember converts members of protobuf to model of member
func DecodeMember(pbmembers ...*chat.Member) (members []*Member) {
	for _, reqMember := range pbmembers {
		members = append(members, &Member{
			Model: gorm.Model{
				ID: uint(reqMember.Id),
			},
			UserID: uint(reqMember.UserId),
			Role:   reqMember.Role.String(),
			Name:   reqMember.Name,
		})
	}
	return
}

func (m *memberRepository) UpdateLastMessageID(memberID uint, messageID uint64) error {
	return m.db.Model(&Member{}).Where("id = ?", memberID).Update("last_message_id", messageID).Error
}
