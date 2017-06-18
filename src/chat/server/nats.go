package server

import (
	"chat/config"
	"chat/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"proto/bot"
	proto "proto/chat"
	"proto/checker"
	"proto/core"
	"strconv"
	"strings"
	"utils/log"
	"utils/mandible"
	"utils/rpc"
)

func (cs *chatServer) handleNotify(notify *bot.Notify) (acknowledged bool) {
	log.Debug("new direct notify: %+v", notify)

	switch {
	// normal notify, not a reply for something
	case notify.ReplyKey == "":

	case strings.HasPrefix(notify.ReplyKey, models.MessageReplyPrefix):
		return cs.handleMessageReply(notify)

	case strings.HasPrefix(notify.ReplyKey, models.ThreadReplyPrefix):
		return cs.handleThreadReply(notify)

	default:
		// do not care about foreign replies
		return true
	}

	chat, err := cs.chats.GetByDirectThread(notify.ThreadId)
	if err != nil {
		log.Errorf("failed to load chat by direct thread %v: %v", notify.ThreadId, err)
		return false
	}

	// unknown/new conversation. ignore it, we should get crate thread reply before any interesting messages
	if chat == nil {
		log.Debug("unknown thread %v", notify.ThreadId)
		return true
	}

	// listen to primary source only
	if notify.SourceId != chat.PrimaryInstagram {
		return true
	}

	for _, msg := range notify.Messages {
		retry, err := cs.handleNewMessage(chat, msg)
		if err != nil {
			log.Error(err)
			if retry {
				return false
			}
		}
	}
	return true
}

func (cs *chatServer) handleNewMessage(chat *models.Conversation, msg *bot.Message) (retry bool, err error) {
	exists, err := cs.chats.CheckMessageExists(msg.MessageId)
	if err != nil {
		return true, fmt.Errorf("failed to check message existence: %v", err)
	}
	if exists {
		log.Debug("message %v already exists", msg.MessageId)
		return false, nil
	}

	// ignore init message
	// @TODO may be there is better way to do it
	if msg.UserId == chat.PrimaryInstagram && msg.Type == bot.MessageType_Text &&
		(msg.Data == chat.Caption || msg.Data == models.DefaultSyncInitMessage) {
		return false, nil
	}

	author, notExists, err := cs.getAuthor(chat, msg.UserId)
	if notExists {
		log.Warn("instagram user with id %v not exists", msg.UserId)
		return false, nil
	}
	if err != nil {
		return true, fmt.Errorf("failed to get message author: %v", err)
	}

	// well, it isn't normal
	if author == nil {
		return true, fmt.Errorf("getAuthor returned nil")
	}

	var parts = make([]*models.MessagePart, 0, 8)

	switch msg.Type {
	case bot.MessageType_Text:
		parts = append(parts, &models.MessagePart{
			Content:  msg.Data,
			MimeType: "text/plain",
		})

	case bot.MessageType_Image:
		img, err := models.ImageUploader.DoRequest("url", msg.Data)
		switch resp := err.(type) {
		case nil:
			j, _ := json.Marshal(img)
			parts = append(parts, &models.MessagePart{
				MimeType:  "image/json",
				ContentID: img.Hash,
				Content:   string(j),
			})

		case *mandible.ImageResp:
			if resp.Status < 400 || resp.Status >= 500 {
				return true, err
			}
			log.Warn("direct notify for message %v contains invalid image", msg.MessageId, msg.Type)
			parts = append(parts, &models.MessagePart{
				Content:  msg.Data,
				MimeType: "text/plain",
			})

		default:
			return true, err
		}

	default:
		log.Warn("direct notify for message %v with unsupported content type %v was skipped", msg.MessageId, msg.Type)
		return false, nil
	}

	_, err = cs.sendMessage(chat, &models.Message{
		MemberID:    sql.NullInt64{Int64: int64(author.ID), Valid: true},
		Member:      author,
		InstagramID: msg.MessageId,
		SyncStatus:  proto.SyncStatus_SYNCED,
		Parts:       parts,
	})
	if err != nil {
		return true, fmt.Errorf("failed to add message: %v", err)
	}
	return false, nil
}

func (cs *chatServer) handleMessageReply(notify *bot.Notify) (acknowledged bool) {
	msgID, err := strconv.ParseUint(strings.TrimPrefix(notify.ReplyKey, models.MessageReplyPrefix), 10, 64)
	if err != nil {
		log.Errorf("bad format of send direct reply key '%v'", notify.ReplyKey)
		return true
	}
	log.Debug("got message send reply for chat %v", msgID)
	// @TODO check what kind of error we have. May be we should handle deleted threads in special way for example
	var (
		status      proto.SyncStatus
		instagramID string
	)
	if notify.Error != "" {
		log.Errorf("error in send direct reply: %v", notify.Error)
		status = proto.SyncStatus_ERROR
	} else {
		status = proto.SyncStatus_SYNCED
		instagramID = notify.Messages[0].MessageId
	}
	err = cs.chats.UpdateSyncStatus(msgID, instagramID, status)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "unique_message_id"` {
			log.Errorf("UpdateSyncStatus: message with instagram_id = %v already exists", notify.Messages[0].MessageId)
			return true
		}
		log.Errorf("UpdateSyncStatus failed: %v", err)
		return false
	}
	return true
}

func (cs *chatServer) handleThreadReply(notify *bot.Notify) (acknowledged bool) {
	chatID, err := strconv.ParseUint(strings.TrimPrefix(notify.ReplyKey, models.ThreadReplyPrefix), 10, 64)
	if err != nil {
		log.Errorf("bad format of create thread reply key '%v'", notify.ReplyKey)
		return true
	}
	log.Debug("got thread create reply for chat %v", chatID)
	if notify.Error != "" {
		log.Errorf("error in create thread reply for chat %v: %v", chatID, notify.Error)
		retry, err := cs.chats.SetSyncError(chatID)
		if err == nil {
			return true
		}
		log.Errorf("failed to set sync error for chat %v: %v", chatID, notify.Error)
		return !retry
	}
	retry, err := cs.chats.SetRelatedThread(chatID, notify.ThreadId)
	if err != nil {
		log.Errorf("failed to set related direct thread: %v", err)
	}
	return !retry
}

func (cs *chatServer) getAuthor(chat *models.Conversation, instagramID uint64) (author *models.Member, notExists bool, err error) {
	for _, member := range chat.Members {
		if member.InstagramID == instagramID {
			author = member
			// prefer supplier, sellers may have same local instagram id due able to answer from shop name
			if author.Role == "SUPPLIER" {
				return
			}
		}
	}
	if author != nil {
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	userReply, err := cs.userCli.ReadUser(ctx, &core.ReadUserRequest{
		InstagramId: instagramID,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to read user: %v", err)
	}

	var user *core.User

	// new user even for core
	if userReply.Id == 0 {
		user, notExists, err = cs.createUser(instagramID)
		if notExists {
			return
		}
		if err != nil {
			return nil, false, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		user = userReply.User
	}

	for _, member := range chat.Members {
		if int64(member.UserID) == user.Id {
			member.InstagramID = instagramID
			err = cs.chats.UpdateMember(member)
			if err != nil {
				return nil, false, fmt.Errorf("failed to update member: %v", err)
			}
			author = member
			return
		}
	}

	ctx, cancel = rpc.DefaultContext()
	defer cancel()
	roleReply, err := cs.leadCli.GetUserRole(ctx, &core.GetUserRoleRequest{
		UserId:         uint64(user.Id),
		ConversationId: chat.ID,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to determinate user role: %v", err)
	}
	if roleReply.Error != "" {
		return nil, false, fmt.Errorf("failed to determinate user role: %v", roleReply.Error)
	}

	name := user.Name
	if name == "" {
		name = user.InstagramUsername
	}
	author = &models.Member{
		ConversationID: chat.ID,
		UserID:         uint64(user.Id),
		Name:           name,
		Role:           roleReply.Role.String(),
		InstagramID:    instagramID,
	}
	// AddMembers will set member ID after save, but i'm still unsure if it is fair usage
	err = cs.chats.AddMembers(chat, author)
	if err != nil {
		return nil, false, fmt.Errorf("fialed to add member to conversation: %v", err)
	}
	return author, false, nil
}

func (cs *chatServer) createUser(instagramID uint64) (user *core.User, notExists bool, err error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	profileReply, err := cs.checkerCli.GetProfile(ctx, &checker.GetProfileRequest{
		Id: instagramID,
	})
	if err != nil {
		return nil, false, err
	}
	switch profileReply.Error {
	case "":
	case "user not found":
		return nil, true, nil
	default:
		return nil, false, errors.New(profileReply.Error)
	}

	avatarURL, _, err := mandible.New(config.Get().UploadService).
		UploadImageByURL(profileReply.AvatarUrl)
	if err != nil {
		return nil, false, fmt.Errorf("failed to upload avatar: %v", err)
	}

	ctx, cancel = rpc.DefaultContext()
	defer cancel()

	userReply, err := cs.userCli.FindOrCreateUser(ctx, &core.CreateUserRequest{
		User: &core.User{
			InstagramId:        instagramID,
			InstagramUsername:  profileReply.Name,
			InstagramFullname:  profileReply.FullName,
			InstagramCaption:   profileReply.Biography,
			InstagramAvatarUrl: profileReply.AvatarUrl,
			AvatarUrl:          avatarURL,
			Website:            profileReply.ExternalUrl,
		},
	})

	if err != nil {
		return nil, false, err
	}

	return userReply.User, false, nil
}

func (cs *chatServer) enableSync(chatID uint64) bool {
	retry, err := cs.chats.EnableSync(chatID, 0, "", false)
	if err != nil {
		log.Errorf("failed to enable sync: %v", err)
	}
	return retry
}
