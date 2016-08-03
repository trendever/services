package server

import (
	"chat/models"
	"chat/publisher"
	"chat/queue"
	"errors"
	"golang.org/x/net/context"
	proto_chat "proto/chat"
)

type chatServer struct {
	chats models.ConversationRepository
	queue queue.Waiter
}

//NewChatServer returns implementation of protobuf ChatServiceServer
func NewChatServer(chats models.ConversationRepository, q queue.Waiter) proto_chat.ChatServiceServer {
	return &chatServer{chats: chats, queue: q}
}

//CreateChat creates new chat
func (cs *chatServer) CreateChat(ctx context.Context, req *proto_chat.NewChatRequest) (*proto_chat.ChatReply, error) {
	if req.Chat == nil {
		return nil, errors.New("Chat is required")
	}
	chat := &models.Conversation{Name: req.Chat.Name}
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
	reply.Error, reply.Messages, err = cs.sendMessage(chat, req.Messages...)
	reply.Chat = chat.Encode()
	if err == nil && reply.Error == nil {
		go cs.notifyChatAboutNewMessage(reply.Chat, reply.Messages)
	}
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

func (cs *chatServer) sendMessage(chat *models.Conversation, newMessages ...*proto_chat.Message) (chatErr *proto_chat.Error, messages []*proto_chat.Message, err error) {
	messages = []*proto_chat.Message{}
	for _, message := range newMessages {

		var member *models.Member
		member, err = cs.chats.GetMember(chat, message.UserId)
		if err != nil {
			return
		}
		if member == nil {
			chatErr = &proto_chat.Error{
				Code:    proto_chat.ErrorCode_FORBIDDEN,
				Message: "User isn't a member",
			}
			return
		}
		msg := models.DecodeMessage(message, member)
		err = cs.chats.AddMessages(chat, msg)
		if err != nil {
			return
		}
		cs.queue.Push(msg)
		messages = append(messages, msg.Encode())
	}
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

// returns user's chats
func (cs *chatServer) GetUserChats(ctx context.Context, req *proto_chat.UserChatsRequest) (reply *proto_chat.UserChatsReply, err error) {
	return nil, errors.New("Deprecated. Don't use this method. Use lead.list instead of this")
	reply = &proto_chat.UserChatsReply{
		Chats: []*proto_chat.Chat{},
	}
	var chats []*models.Conversation
	if chats, err = cs.chats.GetByUserID(uint(req.UserId)); err == nil {
		for _, chat := range chats {
			reply.Chats = append(reply.Chats, chat.Encode())
		}
	}
	return
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
	member.LastMessageID = uint(req.MessageId)

	go cs.notifyChatAboutReadedMessage(chat.Encode(), req.MessageId, req.UserId)

	return
}

func (cs *chatServer) notifyChatAboutNewMessage(chat *proto_chat.Chat, messages []*proto_chat.Message) {
	publisher.Publish(publisher.EventMessage, &proto_chat.NewMessageRequest{
		Chat:     chat,
		Messages: messages,
	})

}

func (cs *chatServer) notifyChatAboutReadedMessage(chat *proto_chat.Chat, message_id, user_id uint64) {
	publisher.Publish(publisher.EventMessageReaded, &proto_chat.MessageReadedRequest{
		Chat:      chat,
		MessageId: message_id,
		UserId:    user_id,
	})

}

func (cs *chatServer) notifyChatAboutNewMember(chat *proto_chat.Chat, member *proto_chat.Member) {
	publisher.Publish(publisher.EventJoin, &proto_chat.NewChatMemberRequest{
		Chat: chat,
		User: member,
	})
}

func (cs *chatServer) GetTotalCountUnread(_ context.Context, req *proto_chat.TotalCountUnreadRequest) (reply *proto_chat.TotalCountUnreadReply, err error) {
	reply = new(proto_chat.TotalCountUnreadReply)
	reply.Count, err = cs.chats.GetTotalUnread(req.UserId)
	return
}
