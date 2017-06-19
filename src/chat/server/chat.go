package server

import (
	"chat/config"
	"chat/models"
	"chat/queue"
	"errors"
	"golang.org/x/net/context"
	proto_chat "proto/chat"
	"proto/checker"
	"proto/core"
	"strings"
	"time"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

// NATS events
const (
	EventJoin            = "chat.member.join"
	EventMessage         = "chat.message.new"
	EventMessageReaded   = "chat.message.readed"
	EventMessageAppended = "chat.message.appended"
)

type chatServer struct {
	chats      models.ConversationRepository
	queue      queue.Waiter
	userCli    core.UserServiceClient
	leadCli    core.LeadServiceClient
	checkerCli checker.CheckerServiceClient
}

//NewChatServer returns implementation of protobuf ChatServiceServer
func NewChatServer(chats models.ConversationRepository, q queue.Waiter) proto_chat.ChatServiceServer {
	conf := config.Get()

	coreConn := rpc.Connect(conf.RPC.Core)
	srv := &chatServer{
		chats:      chats,
		queue:      q,
		userCli:    core.NewUserServiceClient(coreConn),
		leadCli:    core.NewLeadServiceClient(coreConn),
		checkerCli: checker.NewCheckerServiceClient(rpc.Connect(conf.RPC.Checker)),
	}
	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "direct.notify",
		Group:          "chat",
		DurableName:    "chat",
		AckTimeout:     time.Second * 30,
		DecodedHandler: srv.handleNotify,
	}, &nats.StanSubscription{
		Subject:     "chat.conversation.delete",
		Group:       "chat",
		DurableName: "chat",
		DecodedHandler: func(id uint64) bool {
			// @TODO check errors?
			log.Error(chats.DeleteConversation(id))
			return true
		},
	}, &nats.StanSubscription{
		Subject:     "chat.conversation.set_status",
		Group:       "chat",
		DurableName: "chat",
		DecodedHandler: func(req *proto_chat.SetStatusMessage) bool {
			log.Error(chats.SetConversationStatus(req))
			return true
		},
	})
	return srv
}

//CreateChat creates new chat
func (cs *chatServer) CreateChat(ctx context.Context, req *proto_chat.NewChatRequest) (*proto_chat.ChatReply, error) {
	log.Debug("create chat request: %+v", req)
	if req.Chat == nil {
		return nil, errors.New("Chat is required")
	}
	parts := strings.Split(req.Chat.DirectThread, "#")
	chat := &models.Conversation{
		Name:             req.Chat.Name,
		DirectThread:     parts[0],
		PrimaryInstagram: req.PrimaryInstagram,
		Caption:          req.Chat.Caption,
	}
	if err := cs.chats.Create(chat); err != nil {
		return nil, err
	}
	if req.Chat.Members != nil && len(req.Chat.Members) > 0 {
		members := models.DecodeMember(req.Chat.Members...)
		if err := cs.chats.AddMembers(chat, members...); err != nil {
			return nil, err
		}

	}
	return &proto_chat.ChatReply{
		Chat: chat.Encode(),
	}, nil
}

//GetChat returns exists chat
func (cs *chatServer) GetChats(ctx context.Context, req *proto_chat.ChatsRequest) (reply *proto_chat.ChatsReply, err error) {
	var chats models.Conversations
	reply = &proto_chat.ChatsReply{}
	chats, err = cs.chats.GetByIDs(req.Id)
	if err != nil {
		return reply, err
	}
	unread, err := cs.chats.GetUnread(req.Id, req.UserId)
	if err != nil {
		return
	}
	chats.AddUnread(unread)
	reply.Chats = chats.Encode()
	return reply, err

}

//JoinChat add users to the chat
func (cs *chatServer) JoinChat(ctx context.Context, req *proto_chat.JoinChatRequest) (reply *proto_chat.ChatReply, err error) {
	var chat *models.Conversation
	reply = &proto_chat.ChatReply{}
	chat, reply.Error, err = cs.getChat(req.ConversationId)
	if reply.Error != nil || err != nil {
		return reply, err
	}
	members := models.DecodeMember(req.Members...)
	err = cs.chats.AddMembers(chat, members...)
	if err != nil {
		return
	}
	reply.Chat = chat.Encode()
	for _, m := range req.Members {
		mm := chat.GetMember(m.UserId)
		if mm != nil {
			go cs.notifyChatAboutNewMember(reply.Chat, mm.Encode())
		}
	}

	return reply, err
}

//LeaveChat remove users from the chat
func (cs *chatServer) LeaveChat(ctx context.Context, req *proto_chat.LeaveChatRequest) (reply *proto_chat.ChatReply, err error) {
	var chat *models.Conversation
	reply = &proto_chat.ChatReply{}
	chat, reply.Error, err = cs.getChat(req.ConversationId)
	if reply.Error != nil || err != nil {
		return reply, err
	}
	err = cs.chats.RemoveMembers(chat, req.UserIds...)
	reply.Chat = chat.Encode()
	return reply, err
}

//SendNewMessage sends new message from the user to the chat
func (cs *chatServer) SendNewMessage(ctx context.Context, req *proto_chat.SendMessageRequest) (reply *proto_chat.SendMessageReply, err error) {
	reply = &proto_chat.SendMessageReply{}
	var chat *models.Conversation
	chat, reply.Error, err = cs.getChat(req.ConversationId)
	if reply.Error != nil || err != nil {
		return
	}
	msgs := make([]*models.Message, len(req.Messages))
	for i, enc := range req.Messages {
		var member *models.Member
		found := false
		for _, member = range chat.Members {
			if member.UserID == enc.UserId {
				found = true
				break
			}
		}
		if !found {
			reply.Error = &proto_chat.Error{
				Code:    proto_chat.ErrorCode_FORBIDDEN,
				Message: "User isn't a member",
			}
		}
		msgs[i] = models.DecodeMessage(enc, member)
	}

	reply.Messages, err = cs.sendMessage(chat, msgs...)
	reply.Chat = chat.Encode()
	return reply, err
}

func (cs *chatServer) getChat(id uint64) (*models.Conversation, *proto_chat.Error, error) {
	chat, err := cs.chats.GetByID(uint(id))
	if err != nil {
		return nil, nil, err
	}
	var chatErr *proto_chat.Error
	if chat == nil {
		chatErr = &proto_chat.Error{
			Code:    proto_chat.ErrorCode_NOT_EXISTS,
			Message: "Conversation not exists",
		}
	}
	return chat, chatErr, nil
}

func (cs *chatServer) sendMessage(chat *models.Conversation, messages ...*models.Message) (encoded []*proto_chat.Message, err error) {
	err = cs.chats.AddMessages(chat, messages...)
	if err != nil {
		return
	}
	for _, message := range messages {
		cs.queue.Push(message)
		encoded = append(encoded, message.Encode())
	}
	// @TODO wrong place?
	go cs.notifyChatAboutNewMessage(chat.Encode(), encoded)
	return
}

//GetChatHistory returns chat history
func (cs *chatServer) GetChatHistory(ctx context.Context, req *proto_chat.ChatHistoryRequest) (reply *proto_chat.ChatHistoryReply, err error) {
	var chat *models.Conversation
	reply = &proto_chat.ChatHistoryReply{}
	chat, reply.Error, err = cs.getChat(req.ConversationId)
	if err != nil || reply.Error != nil {
		return reply, err
	}
	var member *models.Member
	member, err = cs.chats.GetMember(chat, req.UserId)
	if err != nil {
		return
	}
	if member == nil {
		reply.Error = &proto_chat.Error{
			Code:    proto_chat.ErrorCode_FORBIDDEN,
			Message: "User isn't a member",
		}
		return
	}
	reply.Chat = chat.Encode()
	messages, err := cs.chats.GetHistory(chat, req.FromMessageId, req.Limit, req.Direction)
	if err != nil {
		return
	}
	for _, message := range messages {
		reply.Messages = append(reply.Messages, message.Encode())
	}
	reply.TotalMessages = cs.chats.TotalMessages(chat)
	return reply, err
}

// updated last message id for member
func (cs *chatServer) MarkAsReaded(ctx context.Context, req *proto_chat.MarkAsReadedRequest) (reply *proto_chat.MarkAsReadedReply, err error) {
	reply = &proto_chat.MarkAsReadedReply{}
	var chat *models.Conversation
	chat, reply.Error, err = cs.getChat(req.ConversationId)
	if reply.Error != nil || err != nil {
		return
	}
	var member *models.Member
	member, err = cs.chats.GetMember(chat, req.UserId)
	if err != nil {
		return
	}
	if member == nil {
		err = errors.New("User is not a member")
		return
	}
	err = cs.chats.MarkAsReaded(member, req.MessageId)

	if err != nil {
		return
	}

	//We don't want make a new query only for get updated last_message_id
	//so just update it in the structure
	member = chat.GetMember(req.UserId)
	member.LastMessageID = req.MessageId

	go cs.notifyChatAboutReadedMessage(chat.Encode(), req.MessageId, req.UserId)

	return
}

func (cs *chatServer) AppendMessage(ctx context.Context, req *proto_chat.AppendMessageRequest) (reply *proto_chat.AppendMessageReply, err error) {

	message, err := cs.chats.UpdateMessage(req.MessageId, models.DecodeParts(req.Parts))
	if err != nil {
		return nil, err
	}

	var encMsg = message.Encode()

	go cs.notifyChatAboutAppendedMessage(encMsg)

	return &proto_chat.AppendMessageReply{
		Message: encMsg,
	}, nil
}

func (cs *chatServer) notifyChatAboutAppendedMessage(msg *proto_chat.Message) {

	// api needs chat because it contains users who needs notification about an event
	chat, err := cs.chats.GetByID(uint(msg.ConversationId))
	if err != nil {
		return
	}

	nats.StanPublish(EventMessageAppended, &proto_chat.MessageAppendedRequest{
		Message: msg,
		Chat:    chat.Encode(),
	})
}

func (cs *chatServer) notifyChatAboutNewMessage(chat *proto_chat.Chat, messages []*proto_chat.Message) {
	nats.StanPublish(EventMessage, &proto_chat.NewMessageRequest{
		Chat:     chat,
		Messages: messages,
	})
}

func (cs *chatServer) notifyChatAboutReadedMessage(chat *proto_chat.Chat, messageID, userID uint64) {
	nats.StanPublish(EventMessageReaded, &proto_chat.MessageReadedRequest{
		Chat:      chat,
		MessageId: messageID,
		UserId:    userID,
	})

}

func (cs *chatServer) notifyChatAboutNewMember(chat *proto_chat.Chat, member *proto_chat.Member) {
	nats.StanPublish(EventJoin, &proto_chat.NewChatMemberRequest{
		Chat: chat,
		User: member,
	})
}

func (cs *chatServer) GetTotalCountUnread(_ context.Context, req *proto_chat.TotalCountUnreadRequest) (reply *proto_chat.TotalCountUnreadReply, err error) {
	reply = new(proto_chat.TotalCountUnreadReply)
	reply.Count, err = cs.chats.GetTotalUnread(req.UserId)
	return
}

func (cs *chatServer) EnableSync(_ context.Context, req *proto_chat.EnableSyncRequest) (*proto_chat.EnableSyncReply, error) {
	var reply proto_chat.EnableSyncReply
	parts := strings.Split(req.ThreadId, "#")
	var since string
	if len(parts) > 1 {
		since = parts[1]
	}
	_, err := cs.chats.EnableSync(req.ChatId, req.PrimaryInstagram, parts[0], since, req.ForceNewThread)
	if err != nil {
		reply.Error = err.Error()
	}
	return &reply, err
}
