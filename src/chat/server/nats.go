package server

import (
	"chat/config"
	"chat/models"
	"database/sql"
	"errors"
	"fmt"
	"proto/bot"
	"proto/checker"
	"proto/core"
	"utils/log"
	"utils/mandible"
	"utils/rpc"
)

func (cs *chatServer) handleDirectMessage(notify *bot.DirectMessageNotify) (acknowledged bool) {
	log.Debug("new direct message: %+v", notify)

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

	// unknown/new conversation
	if chat == nil {
		// @TODO should we wait until lead will be created?
		log.Debug("unknown thread %v", notify.ThreadId)
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

	err = cs.chats.AddMessages(chat, &models.Message{
		MemberID:    sql.NullInt64{Int64: int64(author.ID), Valid: true},
		Member:      author,
		InstagramID: notify.MessageId,
		Parts: []*models.MessagePart{
			{
				Content:  notify.Text,
				MimeType: "text/plain",
			},
		},
	})
	if err != nil {
		log.Errorf("failed to add message: %v", err)
		return false
	} else {
		return true
	}
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