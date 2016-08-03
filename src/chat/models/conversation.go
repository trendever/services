package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	pb_chat "proto/chat"
)

//Conversation is representation of conversation model
type Conversation struct {
	gorm.Model
	Name        string
	Members     []*Member
	Messages    []*Message
	Status      string `gorm:"index;default:'new'"`
	UnreadCount uint64 `sql:"-"`
}

type conversationRepositoryImpl struct {
	db       *gorm.DB
	members  *memberRepository
	messages MessageRepository
}

//Conversations is collection of conversation models
type Conversations []*Conversation

//NewConversationRepository returns repository for to work with models of conversation
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepositoryImpl{db: db, members: &memberRepository{db: db}}
}

//ConversationRepository is repository interface of conversation models
type ConversationRepository interface {
	Create(*Conversation) error
	GetByID(uint) (*Conversation, error)
	GetByIDs([]uint64) (Conversations, error)
	GetByUserID(uint) ([]*Conversation, error)
	AddMembers(*Conversation, ...*Member) error
	RemoveMembers(*Conversation, ...uint64) error
	AddMessages(*Conversation, ...*Message) error
	GetMember(*Conversation, uint64) (*Member, error)
	GetHistory(chat *Conversation, fromMessageID uint64, limit uint64, direction bool) ([]*Message, error)
	TotalMessages(chat *Conversation) uint64
	MarkAsReaded(member *Member, messageID uint64) error
	GetUnread(ids []uint64, userID uint64) (map[uint]uint64, error)
	GetTotalUnread(userID uint64) (uint64, error)
}

//Encode converts to protobuf model
func (c *Conversation) Encode() *pb_chat.Chat {
	chat := &pb_chat.Chat{}
	chat.Id = uint64(c.ID)
	chat.Name = c.Name
	chat.UnreadCount = c.UnreadCount
	if c.Members != nil {
		chat.Members = []*pb_chat.Member{}
		for _, m := range c.Members {
			chat.Members = append(chat.Members, m.Encode())
		}
	}
	if c.Messages != nil && len(c.Messages) == 1 {
		chat.RecentMessage = c.Messages[0].Encode()
	}
	return chat
}

func (c *conversationRepositoryImpl) Create(model *Conversation) error {
	return c.db.Create(model).Error
}

func (c *conversationRepositoryImpl) GetByID(id uint) (model *Conversation, err error) {
	model = &Conversation{}
	scope := c.defaultPreload([]uint64{uint64(id)}).Find(model, id)
	if scope.RecordNotFound() {
		return nil, nil
	}
	err = scope.Error
	return
}

func (c *conversationRepositoryImpl) AddMembers(chat *Conversation, members ...*Member) error {

	for _, member := range members {
		exists, err := c.GetMember(chat, uint64(member.UserID))
		if err != nil {
			return err
		}
		if exists == nil {
			err := c.db.Model(chat).Association("Members").Append(member).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *conversationRepositoryImpl) RemoveMembers(chat *Conversation, userIDs ...uint64) error {
	for _, userID := range userIDs {
		member, err := c.GetMember(chat, userID)
		if err != nil {
			return err
		}
		if member == nil {
			continue
		}
		if err := c.db.Delete(member).Error; err != nil {
			return err
		}
	}
	return nil
}

func (c *conversationRepositoryImpl) AddMessages(chat *Conversation, messages ...*Message) error {
	if chat.Status == "new" {
		for _, m := range messages {
			if m.Member != nil && m.Member.Role != pb_chat.MemberRole_name[int32(pb_chat.MemberRole_SYSTEM)] {
				chat.Status = "active"
				err := c.db.Save(chat).Error
				if err != nil {
					return fmt.Errorf("failed to update chat status: %v", err)
				}
			}
		}
	}
	return c.db.Model(chat).Association("Messages").Append(messages).Error
}

func (c *conversationRepositoryImpl) GetMember(model *Conversation, userID uint64) (member *Member, err error) {
	member = &Member{}
	scope := c.db.Where("user_id = ? AND conversation_id = ?", userID, model.ID).Find(member)
	if scope.RecordNotFound() {
		return nil, nil
	}
	err = scope.Error
	return
}

func (c *conversationRepositoryImpl) GetHistory(chat *Conversation, fromMessageID uint64, limit uint64, direction bool) (messages []*Message, err error) {
	messages = []*Message{}
	scope := c.db.
		Preload("Parts").
		Preload("Member").
		Model(&Message{}).
		Where("conversation_id = ?", chat.ID)
		//Order("created_at desc")
	if fromMessageID > 0 {
		if direction {
			scope = scope.Where("id > ?", fromMessageID)
		} else {
			scope = scope.Where("id < ?", fromMessageID)
		}
	}
	if direction {
		scope = scope.Order("created_at desc")
	} else {
		scope = scope.Order("created_at asc")
	}
	if limit > 0 {
		scope = scope.Limit(int(limit))
	} else {
		scope = scope.Limit(20)
	}
	err = scope.Find(&messages).Error
	return
}

func (c *conversationRepositoryImpl) TotalMessages(chat *Conversation) uint64 {
	return uint64(c.db.Model(chat).Association("Messages").Count())
}

func (c *conversationRepositoryImpl) GetByUserID(userID uint) ([]*Conversation, error) {
	rows, err := c.db.Model(Member{}).Where("user_id = ?", userID).Select("conversation_id").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint
	for rows.Next() {
		var id uint
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	var chats = []*Conversation{}
	if len(ids) == 0 {
		return chats, nil
	}
	err = c.db.Preload("Members").Find(&chats, "id in (?)", ids).Error
	return chats, err
}

func (c *conversationRepositoryImpl) MarkAsReaded(member *Member, messageID uint64) error {
	return c.members.UpdateLastMessageID(member.ID, messageID)
}

//GetByIDs returns conversations with members and last messages
func (c *conversationRepositoryImpl) GetByIDs(ids []uint64) (models Conversations, _ error) {
	scope := c.defaultPreload(ids).
		Find(&models, "id IN (?)", ids)

	return models, scope.Error
}

func (c *conversationRepositoryImpl) defaultPreload(ids []uint64) *gorm.DB {
	return c.db.
		Preload("Members").
		Preload("Messages", "id IN (SELECT MAX(id) FROM messages WHERE conversation_id in (?) GROUP BY conversation_id)", ids).
		Preload("Messages.Parts").
		Preload("Messages.Member")
}

//GetUnread returns count of unread messages mapped to conversation ids
func (c *conversationRepositoryImpl) GetUnread(ids []uint64, userID uint64) (map[uint]uint64, error) {
	unreadMap := map[uint]uint64{}
	rows, err := c.db.Model(&Message{}).
		Select("count(messages.id), messages.conversation_id").
		Joins("LEFT JOIN members ON (members.conversation_id = messages.conversation_id AND members.user_id = ?)", userID).
		Where("(messages.id > members.last_message_id OR members.last_message_id IS NULL)").
		Where("messages.conversation_id in (?)", ids).
		Group("messages.conversation_id").
		Rows()
	if err != nil {
		return unreadMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var count uint64
		err := rows.Scan(&count, &id)
		if err != nil {
			return unreadMap, err
		}
		unreadMap[id] = count
	}
	return unreadMap, nil
}

func (c *conversationRepositoryImpl) GetTotalUnread(userID uint64) (uint64, error) {
	var missed uint64
	err := c.db.
		Select("COUNT(DISTINCT c.id)").
		Table("members u").
		Joins("JOIN conversations c ON u.conversation_id = c.id").
		Joins("JOIN messages m ON m.conversation_id = c.id").
		Where("u.user_id = ?", userID).
		Where("u.last_message_id < m.id").
		Where("u.role = 'CUSTOMER' OR (u.role != 'CUSTOMER' AND c.status != 'new')").
		Row().
		Scan(&missed)
	return missed, err
}

//Encode converts to protobuf model
func (c Conversations) Encode() (chats []*pb_chat.Chat) {
	for _, ch := range c {
		chats = append(chats, ch.Encode())
	}
	return
}

//AddUnread adds unread count
func (c Conversations) AddUnread(unread map[uint]uint64) {
	for _, ch := range c {
		ch.UnreadCount = 0
		count, ok := unread[ch.ID]
		if ok {
			ch.UnreadCount = count
		}
	}
}

func (c *Conversation) GetMember(user_id uint64) *Member {
	if c.Members == nil {
		return nil
	}
	for _, m := range c.Members {
		if m.UserID == uint(user_id) {
			return m
		}
	}
	return nil
}
