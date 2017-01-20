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

func (cs *chatServer) handleDirectNotify(notify *bot.DirectNotify) (acknowledged bool) {
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

	exists, err := cs.chats.CheckMessageExists(notify.MessageId)
	if err != nil {
		log.Errorf("failed to check message existence: %v", err)
		return false
	}
	if exists {
		log.Debug("message %v already exists", notify.MessageId)
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

	// ignore init message
	// @TODO may be there is better way to do it
	if notify.UserId == chat.PrimaryInstagram && notify.Type == bot.MessageType_Text &&
		(notify.Data == chat.Caption || notify.Data == models.DefaultSyncInitMessage) {
		return true
	}

	author, err := cs.getAuthor(chat, notify.UserId)
	if err != nil {
		log.Errorf("failed to get message author: %v", err)
		return false
	}

	// well, it isn't normal
	if author == nil {
		log.Errorf("getAuthor returned nil")
		return false
	}

	var parts = make([]*models.MessagePart, 0, 8)
	switch notify.Type {
	case bot.MessageType_Text:
		parts = append(parts, &models.MessagePart{
			Content:  notify.Data,
			MimeType: "text/plain",
		})
	case bot.MessageType_Image:
		img, err := models.ImageUploader.DoRequest("url", notify.Data)
		switch resp := err.(type) {
		case nil:

		case *mandible.ImageResp:
			if resp.Status < 400 || resp.Status >= 500 {
				return false
			}
			log.Warn("direct notify for message %v contains invalid image", notify.MessageId, notify.Type)
			return true

		default:
			return false
		}
		j, _ := json.Marshal(img)
		parts = append(parts, &models.MessagePart{
			MimeType:  "image/json",
			ContentID: img.Hash,
			Content:   string(j),
		})

	default:
		log.Warn("direct notify for message %v with unsupported content type %v was skipped", notify.MessageId, notify.Type)
		return true
	}

	_, err = cs.sendMessage(chat, &models.Message{
		MemberID:    sql.NullInt64{Int64: int64(author.ID), Valid: true},
		Member:      author,
		InstagramID: notify.MessageId,
		SyncStatus:  proto.SyncStatus_SYNCED,
		Parts:       parts,
	})
	if err != nil {
		log.Errorf("failed to add message: %v", err)
		return false
	} else {
		return true
	}
}

func (cs *chatServer) handleMessageReply(notify *bot.DirectNotify) (acknowledged bool) {
	msgID, err := strconv.ParseUint(strings.TrimPrefix(notify.ReplyKey, models.MessageReplyPrefix), 10, 64)
	if err != nil {
		log.Errorf("bad format of send direct reply key '%v'", notify.ReplyKey)
		return true
	}
	log.Debug("got message send reply for chat %v", msgID)
	// @TODO check what kind of error we have. May be we should handle deleted threads in special way for example
	status := proto.SyncStatus_SYNCED
	if notify.Error != "" {
		log.Errorf("error in send direct reply: %v", notify.Error)
		status = proto.SyncStatus_ERROR
	}
	err = cs.chats.UpdateSyncStatus(msgID, notify.MessageId, status)
	if err != nil {
		log.Errorf("UpdateSyncStatus failed: %v", err)
		return false
	}
	return true
}

func (cs *chatServer) handleThreadReply(notify *bot.DirectNotify) (acknowledged bool) {
	chatID, err := strconv.ParseUint(strings.TrimPrefix(notify.ReplyKey, models.ThreadReplyPrefix), 10, 64)
	if err != nil {
		log.Errorf("bad format of create thread reply key '%v'", notify.ReplyKey)
		return true
	}
	log.Debug("got thread create reply for chat %v", chatID)
	if notify.Error != "" {
		log.Errorf("error in create thread reply for chat %v: %v", chatID, notify.Error)
		// @TODO anything else?
		return true
	}
	retry, err := cs.chats.SetRelatedThread(chatID, notify.ThreadId)
	if err != nil {
		log.Errorf("failed to set related direct thread: %v", err)
	}
	return !retry
}

func (cs *chatServer) getAuthor(chat *models.Conversation, instagramID uint64) (author *models.Member, err error) {
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
		return nil, fmt.Errorf("failed to read user: %v", err)
	}

	var user *core.User

	// new user even for core
	if userReply.Id == 0 {
		user, err = cs.createUser(instagramID)
		if err != nil {
			return nil, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		user = userReply.User
	}

	for _, member := range chat.Members {
		if int64(member.UserID) == user.Id {
			member.InstagramID = instagramID
			err = cs.chats.UpdateMember(member)
			if err != nil {
				return nil, fmt.Errorf("failed to update member: %v", err)
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
		return nil, fmt.Errorf("failed to determinate user role: %v", err)
	}
	if roleReply.Error != "" {
		return nil, fmt.Errorf("failed to determinate user role: %v", roleReply.Error)
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
		return nil, fmt.Errorf("fialed to add member to conversation: %v", err)
	}
	return author, nil
}

func (cs *chatServer) createUser(instagramID uint64) (*core.User, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	profileReply, err := cs.checkerCli.GetProfile(ctx, &checker.GetProfileRequest{
		Id: instagramID,
	})
	if err != nil {
		return nil, err
	}
	if profileReply.Error != "" {
		return nil, errors.New(profileReply.Error)
	}

	avatarURL, _, err := mandible.New(config.Get().UploadService).
		UploadImageByURL(profileReply.AvatarUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %v", err)
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
		return nil, err
	}

	return userReply.User, nil
}

func (cs *chatServer) enableSync(chatID uint64) bool {
	retry, err := cs.chats.EnableSync(chatID, 0, "", false)
	if err != nil {
		log.Errorf("failed to enable sync: %v", err)
	}
	return retry
}
