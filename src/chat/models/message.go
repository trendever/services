package models

import (
	"common/db"
	"common/log"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"proto/chat"
	"utils/mandible"
)

//Message is model of message
type Message struct {
	db.Model
	ConversationID uint
	InstagramID    string
	UserID         uint64
	Member         Member `gorm:"ForeignKey:ConversationID,UserID;AssociationForeignKey:ConversationID,UserID"`
	SyncStatus     chat.SyncStatus
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

var ImageUploader *mandible.Uploader

func InitUploader(mandibleUrl string) {
	ImageUploader = mandible.New(mandibleUrl, mandible.Thumbnail{
		Name:   "big",
		Width:  1080,
		Height: 1080,
		Shape:  "thumb",
	}, mandible.Thumbnail{
		Name:   "small",
		Width:  480,
		Height: 480,
		Shape:  "thumb",
	}, mandible.Thumbnail{
		Name:   "small_crop",
		Width:  480,
		Height: 480,
		Shape:  "square",
	}, mandible.Thumbnail{
		Name:  "instagram",
		Shape: "instagram",
	})
}

//Encode converts message to protobuf model
func (m *Message) Encode() *chat.Message {

	message := &chat.Message{
		Id:             uint64(m.ID),
		ConversationId: uint64(m.ConversationID),
		UserId:         m.UserID,
		Parts:          m.EncodeParts(),
		CreatedAt:      m.CreatedAt.Unix(),
		SyncStatus:     m.SyncStatus,
	}
	if m.Member.UserID != 0 {
		message.User = m.Member.Encode()
	}
	return message
}

func EncodeMessages(msgs []*Message) (ret []*chat.Message) {
	for _, msg := range msgs {
		ret = append(ret, msg.Encode())
	}
	return
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
		UserID: member.UserID,
		Member: *member,
		Parts:  DecodeParts(pbMessage.Parts),
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
		img, err := ImageUploader.DoRequest("base64", mp.Content)
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

// Sorting shiet
type messageParts []*MessagePart

func (a messageParts) Len() int           { return len(a) }
func (a messageParts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a messageParts) Less(i, j int) bool { return a[i].ID < a[j].ID }
