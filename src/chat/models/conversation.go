package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/bot"
	pb_chat "proto/chat"
	"strconv"
	"strings"
	"sync"
	"utils/db"
	"utils/log"
	"utils/mandible"
	"utils/nats"
)

const (
	MessageReplyPrefix = "sync_msg."
	ThreadReplyPrefix  = "sync_thread."

	DefaultSyncInitMessage = "direct sync enabled"

	SyncEventTopic = "chat.sync_event"
)

var global struct {
	syncLock    sync.Mutex
	threadsLock sync.Mutex
}

//Conversation is representation of conversation model
type Conversation struct {
	db.Model
	Name         string
	Members      []*Member
	Messages     []*Message
	Caption      string `gorm:"text"`
	Status       string `gorm:"index;default:'new'"`
	UnreadCount  uint64 `sql:"-"`
	SyncStatus   pb_chat.SyncStatus
	DirectThread string `gorm:"index"`
	// instagram id of supplier
	PrimaryInstagram uint64
}

type conversationRepositoryImpl struct {
	db      *gorm.DB
	members *memberRepository
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
	GetByDirectThread(string) (*Conversation, error)
	AddMembers(*Conversation, ...*Member) error
	RemoveMembers(*Conversation, ...uint64) error
	AddMessages(*Conversation, ...*Message) error
	GetMember(*Conversation, uint64) (*Member, error)
	UpdateMember(member *Member) error
	GetHistory(chat *Conversation, fromMessageID uint64, limit uint64, direction bool) ([]*Message, error)
	TotalMessages(chat *Conversation) uint64
	MarkAsReaded(member *Member, messageID uint64) error
	GetUnread(ids []uint64, userID uint64) (map[uint64]uint64, error)
	GetTotalUnread(userID uint64) (uint64, error)
	UpdateMessage(messageID uint64, append []*MessagePart) (*Message, error)
	DeleteConversation(id uint64) error
	SetConversationStatus(req *pb_chat.SetStatusMessage) error
	CheckMessageExists(instagramID string) (bool, error)
	EnableSync(chatID, primaryInstagram uint64, threadID string, forceNowThread bool) (retry bool, err error)
	SetRelatedThread(chatID uint64, directThread string) (retry bool, err error)
	UpdateSyncStatus(localID uint64, instagramID string, status pb_chat.SyncStatus) error
}

//Encode converts to protobuf model
func (c *Conversation) Encode() *pb_chat.Chat {
	chat := &pb_chat.Chat{}
	chat.Id = uint64(c.ID)
	chat.Name = c.Name
	chat.UnreadCount = c.UnreadCount
	chat.DirectThread = c.DirectThread
	chat.SyncStatus = c.SyncStatus
	chat.Caption = c.Caption
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

func (c *conversationRepositoryImpl) GetByDirectThread(id string) (*Conversation, error) {
	var conv Conversation
	res := c.db.Preload("Members").Where("direct_thread = ?", id).First(&conv)
	if res.RecordNotFound() {
		return nil, nil
	}
	return &conv, res.Error
}

func (c *conversationRepositoryImpl) AddMembers(chat *Conversation, members ...*Member) error {

	for _, member := range members {
		err := c.db.Model(chat).Association("Members").Append(member).Error
		if err != nil && err.Error() != `pq: duplicate key value violates unique constraint "once_per_conv"` {
			return err
		}
	}
	return nil
}

func (c *conversationRepositoryImpl) UpdateMember(member *Member) error {
	return c.db.Save(member).Error
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
	global.syncLock.Lock()
	err := c.db.Model(chat).Association("Messages").Append(messages).Error
	go func() {
		if err == nil {
			switch chat.SyncStatus {
			case pb_chat.SyncStatus_SYNCED:
				c.syncMessages(chat, messages...)
			case pb_chat.SyncStatus_DETACHED:
				for _, msg := range messages {
					if msg.InstagramID != "" {
						continue
					}
					kind, data := mapToInstagram(chat, msg)
					if kind != bot.MessageType_None && data != "" {
						c.EnableSync(chat.ID, 0, "", true)
						break
					}
				}
			}
		}
		// yes, it is fine to unlock in new gorutine(well, it's allowed at least)
		global.syncLock.Unlock()
	}()
	return err
}

func (c *conversationRepositoryImpl) syncMessages(chat *Conversation, messages ...*Message) {
	var ids []uint64
	if chat.DirectThread == "" {
		log.Errorf("syncMessages called for chat without related direct thread")
		return
	}
	var err error
	for _, msg := range messages {
		if msg.InstagramID != "" {
			continue
		}
		kind, data := mapToInstagram(chat, msg)
		if kind != bot.MessageType_None && data != "" {
			var req = bot.SendDirectRequest{
				SenderId: chat.PrimaryInstagram,
				ThreadId: chat.DirectThread,
				ReplyKey: MessageReplyPrefix + strconv.FormatUint(msg.ID, 10),
				Type:     kind,
				Data:     data,
			}
			log.Debug("send direct request: %+v", req)
			err = nats.StanPublish("direct.send", &req)
			if err != nil {
				log.Errorf("failed to send send direct request via nats: %v", err)
				break
			}
		}
		ids = append(ids, msg.ID)
	}
	// @TODO db errors should be rare but state may become inconsistent... not to much can be done quick btw
	if err != nil {
		err := c.updateSyncStatus(chat, pb_chat.SyncStatus_ERROR)
		if err != nil {
			log.Errorf("failed to update sync status of chat: %v", err)
		}
	}
	err = c.db.Model(&Message{}).Where("id IN (?)", ids).UpdateColumn("sync_status", pb_chat.SyncStatus_PENDING).Error
	if err != nil {
		log.Errorf("failed to update chat info: %v", err)
	}
}

// set passed status, save chat and send notification via stan
func (c *conversationRepositoryImpl) updateSyncStatus(chat *Conversation, status pb_chat.SyncStatus) error {
	chat.SyncStatus = status
	err := c.db.Save(chat).Error
	if err != nil {
		return err
	}
	return nats.StanPublish(SyncEventTopic, chat.Encode())
}

func mapToInstagram(chat *Conversation, message *Message) (kind bot.MessageType, data string) {
	citation := false
	if message.Member.Role == "CUSTOMER" || message.Member.Role == "UNKNOWN" {
		citation = true
	}

	kind = bot.MessageType_Text
	for _, part := range message.Parts {
		switch part.MimeType {
		case "text/plain":
			trimmed := strings.Trim(part.Content, " \t\r\n")
			if trimmed == "" {
				continue
			}
			if citation {
				data += ">"
			}
			data += part.Content + "\n"
		// @TODO @REFACTOR that is definitely ugly method to add media share data
		case "text/data":
			slice := strings.Split(part.Content, "~")
			if len(slice) >= 3 {
				kind = bot.MessageType_MediaShare
				data = slice[2]
				return
			}
		case "image/json":
			var img mandible.Image
			err := json.Unmarshal([]byte(part.Content), &img)
			if err != nil {
				log.Errorf("failed to unmarshel image json in message %v: %v", message.ID, err)
				continue
			}
			if url, ok := img.Thumbs["instagram"]; ok {
				kind = bot.MessageType_Image
				data = url
				return
			}
		}
	}
	data = strings.Trim(data, " \t\r\n")
	return
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

func (c *conversationRepositoryImpl) GetHistory(chat *Conversation, fromMessageID uint64, limit uint64, direction bool) ([]*Message, error) {

	messages := []*Message{}
	scope := c.db.
		Preload("Parts", func(db *gorm.DB) *gorm.DB { return db.Order("id asc") }). // force sorting of parts by id
		Preload("Member").
		Model(&Message{}).
		Where("conversation_id = ?", chat.ID)

	if fromMessageID > 0 {
		if direction { // if true -- from new to old
			scope = scope.Where("id < ?", fromMessageID)
		} else {
			scope = scope.Where("id > ?", fromMessageID)
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

	err := scope.Find(&messages).Error
	return messages, err
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

// UpdateMessage appends new message part to given message; returns it
func (c *conversationRepositoryImpl) UpdateMessage(messageID uint64, parts []*MessagePart) (*Message, error) {

	var message Message

	// find message
	err := c.db.
		Preload("Parts").
		Preload("Member").
		Model(&Message{}).
		Where("id = ?", messageID).
		Find(&message).
		Error

	if err != nil {
		return nil, err
	}

	message.Parts = append(message.Parts, parts...)

	log.Debug("ehohoh %#v", message)

	err = c.db.Save(&message).Error
	if err != nil {
		return nil, err
	}

	return &message, nil
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
func (c *conversationRepositoryImpl) GetUnread(ids []uint64, userID uint64) (map[uint64]uint64, error) {
	unreadMap := map[uint64]uint64{}
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
		var id, count uint64
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
		Joins("JOIN conversations c ON u.conversation_id = c.id AND c.deleted_at IS NULL").
		Joins("JOIN messages m ON m.conversation_id = c.id").
		Where("u.user_id = ?", userID).
		Where("u.last_message_id < m.id").
		Where("c.status != 'cancelled'").
		Where("u.role = 'CUSTOMER' OR c.status != 'new'").
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
func (c Conversations) AddUnread(unread map[uint64]uint64) {
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
		if m.UserID == user_id {
			return m
		}
	}
	return nil
}

func (c *conversationRepositoryImpl) DeleteConversation(id uint64) error {
	return c.db.Where("id = ?", id).Delete(&Conversation{}).Error
}

func (c *conversationRepositoryImpl) SetConversationStatus(req *pb_chat.SetStatusMessage) error {
	return c.db.Model(&Conversation{}).Where("id = ?", req.ConversationId).UpdateColumn("status", req.Status).Error
}

func (c *conversationRepositoryImpl) CheckMessageExists(instagramID string) (bool, error) {
	var count int
	err := db.New().Model(&Message{}).Where("instagram_id = ?", instagramID).Count(&count).Error
	return count != 0, err
}

func (c *conversationRepositoryImpl) EnableSync(chatID, primaryInstagram uint64, threadID string, forceNowThread bool) (retry bool, err error) {
	log.Debug("enabling sync for chat %v...", chatID)

	var chat Conversation
	err = c.db.Preload("Members").First(&chat, "id = ?", chatID).Error
	if err != nil {
		return true, fmt.Errorf("failed to load chat: %v", err)
	}

	if primaryInstagram != 0 && primaryInstagram != chat.PrimaryInstagram {
		chat.PrimaryInstagram = primaryInstagram
		chat.DirectThread = ""
		err := c.updateSyncStatus(&chat, pb_chat.SyncStatus_NONE)
		if err != nil {
			return true, fmt.Errorf("failed to update chat info: %v", err)
		}
	}

	if chat.PrimaryInstagram == 0 {
		return false, fmt.Errorf("chat %v has no primary instagram", chat.ID)
	}

	if threadID != "" && threadID != chat.DirectThread {
		return c.SetRelatedThread(chatID, threadID)
	}

	if forceNowThread {
		chat.DirectThread = ""
		err := c.updateSyncStatus(&chat, pb_chat.SyncStatus_DETACHED)
		if err != nil {
			return true, fmt.Errorf("failed to update chat info: %v", err)
		}
	}

	switch chat.SyncStatus {
	case pb_chat.SyncStatus_SYNCED, pb_chat.SyncStatus_PENDING:
		// nothing to do yet
		return false, nil

	case pb_chat.SyncStatus_NONE, pb_chat.SyncStatus_DETACHED:
		request := bot.CreateThreadRequest{
			Inviter:     chat.PrimaryInstagram,
			InitMessage: DefaultSyncInitMessage,
			ReplyKey:    ThreadReplyPrefix + strconv.FormatUint(chatID, 10),
		}
		if chat.Caption != "" {
			request.Caption = chat.Caption
			request.InitMessage = chat.Caption
		}

		for _, mmb := range chat.Members {
			if mmb.InstagramID != 0 && mmb.InstagramID != chat.PrimaryInstagram {
				request.Participant = append(request.Participant, mmb.InstagramID)
			}
		}
		if len(request.Participant) == 0 {
			return false, fmt.Errorf("chat %v has no participants with known instagram id", chatID)
		}

		log.Debug("create_thread request: %+v", request)

		err = nats.StanPublish("direct.create_thread", &request)
		if err != nil {
			return true, fmt.Errorf("failed to send create_thread request: %v", err)
		}

		err = c.updateSyncStatus(&chat, pb_chat.SyncStatus_PENDING)
		if err != nil {
			return true, fmt.Errorf("failed to update chat info: %v", err)
		}

		return false, nil

	case pb_chat.SyncStatus_ERROR:
		err = c.updateSyncStatus(&chat, pb_chat.SyncStatus_SYNCED)
		if err != nil {
			return true, fmt.Errorf("failed to update chat info: %v", err)
		}
		c.syncRecent(&chat)
		return false, nil
	}
	err = errors.New("unreachable point is reached in EnableSync()")
	log.Error(err)
	return false, err
}

func (c *conversationRepositoryImpl) SetRelatedThread(chatID uint64, directThread string) (retry bool, err error) {
	global.threadsLock.Lock()
	defer global.threadsLock.Unlock()

	var chat Conversation
	scope := c.db.Preload("Members").First(&chat, chatID)
	if scope.RecordNotFound() {
		return false, fmt.Errorf("unknown chat '%v'", chatID)
	}
	if scope.Error != nil {
		return true, fmt.Errorf("failed to load chat: %v", err)
	}

	var old Conversation
	scope = c.db.
		Where("direct_thread = ?", directThread).
		Where("id != ?", chatID).
		Preload("Members").
		First(&old)
	switch {
	case scope.RecordNotFound():

	case scope.Error != nil:
		return true, fmt.Errorf("failed to detach other chats: %v", err)

	default:
		old.DirectThread = ""
		err := c.updateSyncStatus(&old, pb_chat.SyncStatus_DETACHED)
		if err != nil {
			return true, fmt.Errorf("failed to detach other chats: %v", err)
		}
	}

	if chat.DirectThread != "" {
		log.Warn("chat %v already had related instagram thread '%v', replacing", chatID, chat.DirectThread)
	}
	chat.DirectThread = directThread
	err = c.updateSyncStatus(&chat, pb_chat.SyncStatus_SYNCED)
	if err != nil {
		return true, fmt.Errorf("failed to save chat info: %v", err)
	}
	go c.syncRecent(&chat)
	return false, nil
}

func (c *conversationRepositoryImpl) syncRecent(chat *Conversation) {
	global.syncLock.Lock()

	var messages []*Message
	// @TODO any limits?
	err := c.db.
		Where("conversation_id = ?", chat.ID).
		Where("sync_status IN (?)", []pb_chat.SyncStatus{pb_chat.SyncStatus_NONE, pb_chat.SyncStatus_ERROR}).
		Order("id").
		Preload("Parts").Preload("Member").
		Find(&messages).Error
	if err != nil {
		log.Errorf("failed to load recent messages: %v", err)
	} else {
		c.syncMessages(chat, messages...)
	}

	global.syncLock.Unlock()
}

func (c *conversationRepositoryImpl) UpdateSyncStatus(localID uint64, instagramID string, status pb_chat.SyncStatus) error {
	// @TODO what if related thread will be changed after sending sync request? we can get error status on valid thread rarely
	if status == pb_chat.SyncStatus_ERROR {
		// disable synchronization
		var chat Conversation
		scope := c.db.Preload("Members").
			Where("id IN (SELECT conversation_id FROM messages WHERE id = ?)", localID).
			First(&chat)
		if scope.RecordNotFound() {
			return fmt.Errorf("chat with message %v not found", localID)
		}
		if scope.Error != nil {
			return fmt.Errorf("failed to load chat: %v", scope.Error)
		}

		err := c.updateSyncStatus(&chat, pb_chat.SyncStatus_ERROR)
		if err != nil {
			return fmt.Errorf("failed to save chat info: %v", err)
		}
	}
	return c.db.Model(&Message{}).Where("id = ?", localID).UpdateColumns(Message{InstagramID: instagramID, SyncStatus: status}).Error
}
