package models

import (
	"chat/images"
	"database/sql"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"utils/log"
)

//Message is model of message
type Message struct {
	gorm.Model
	ConversationID uint
	MemberID       sql.NullInt64
	Member         *Member
	Parts          []*MessagePart
}

//MessagePart is model of part of message
type MessagePart struct {
	gorm.Model
	Content   string
	ContentID string
	MimeType  string
	MessageID uint
}

//Encode converts message to protobuf model
func (m *Message) Encode() *chat.Message {

	message := &chat.Message{
		Id:             uint64(m.ID),
		ConversationId: uint64(m.ConversationID),
		UserId:         uint64(m.MemberID.Int64),
		Parts:          m.EncodeParts(),
		CreatedAt:      m.CreatedAt.Unix(),
	}
	if m.Member != nil {
		message.User = m.Member.Encode()
	}
	return message
}

//EncodeParts converts MessageParts to protobuf model
func (m *Message) EncodeParts() []*chat.MessagePart {
	parts := []*chat.MessagePart{}
	for _, part := range m.Parts {
		parts = append(parts, &chat.MessagePart{
			Content:   part.Content,
			ContentId: part.ContentID,
			MimeType:  part.MimeType,
		})
	}
	return parts
}

//DecodeMessage creates message from protobuf model
func DecodeMessage(pbMessage *chat.Message, member *Member) *Message {
	message := &Message{
		MemberID: sql.NullInt64{Int64: int64(member.ID), Valid: member.ID != 0},
		Member:   member,
		Parts:    DecodeParts(pbMessage.Parts),
	}
	return message
}

//DecodeParts decodes parts slice from protobuf model
func DecodeParts(parts []*chat.MessagePart) []*MessagePart {

	out := make([]*MessagePart, len(parts))

	for i, pbPart := range parts {
		out[i] = &MessagePart{
			Content:   pbPart.Content,
			ContentID: pbPart.ContentId,
			MimeType:  pbPart.MimeType,
		}
	}

	return out
}

// BeforeSave hook
func (mp *MessagePart) BeforeSave() error {
	switch mp.MimeType {
	case "image/base64":
		img, err := images.UploadBase64(mp.Content)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debug("Image uploaded: %s", img.Link)
		mp.MimeType = "image/json"
		mp.ContentID = img.Hash
		j, _ := json.Marshal(img)
		mp.Content = string(j)
	}
	return nil
}

// IsStatusMessage check is this message is status
func (m *Message) IsStatusMessage() bool {
	for _, p := range m.Parts {
		if p.MimeType == "json/status" {
			return true
		}
	}
	return false
}
