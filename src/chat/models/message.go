package models

import (
	"database/sql"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"utils/log"
	"chat/images"
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

//MessageRepository is messages repository interface
type MessageRepository interface {
	Create(*Message) error
}

//Encode converts message to protobuf model
func (m *Message) Encode() *chat.Message {

	message := &chat.Message{
		Id:             uint64(m.ID),
		ConversationId: uint64(m.ConversationID),
		UserId:         uint64(m.MemberID.Int64),
		Parts:          m.encodeParts(),
		CreatedAt:      m.CreatedAt.Unix(),
	}
	if m.Member != nil {
		message.User = m.Member.Encode()
	}
	return message
}

//encodeParts converts MessageParts to protobuf model
func (m *Message) encodeParts() []*chat.MessagePart {
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
		Parts:    []*MessagePart{},
	}
	for _, pbPart := range pbMessage.Parts {
		message.Parts = append(message.Parts, decodeParts(pbPart))
	}
	return message
}

//NewPartFromPB creates message part from protobuf model
func decodeParts(pbPart *chat.MessagePart) *MessagePart {
	return &MessagePart{
		Content:   pbPart.Content,
		ContentID: pbPart.ContentId,
		MimeType:  pbPart.MimeType,
	}
}

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

func (m *Message) IsStatusMessage() bool {
	for _, p := range m.Parts {
		if p.MimeType == "json/status" {
			return true
		}
	}
	return false
}
