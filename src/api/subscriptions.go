package main

import (
	"api/cache"
	schat "api/chat"
	"api/soso"
	"proto/chat"
	"utils/nats"
)

func init() {
	nats.Subscribe(&nats.Subscription{
		Subject: "chat.message.new",
		Handler: newMessage,
	}, &nats.Subscription{
		Subject: "chat.message.readed",
		Handler: messageReaded,
	}, &nats.Subscription{
		Subject: "chat.message.appended",
		Handler: messageAppended,
	}, &nats.Subscription{
		Subject: "chat.member.join",
		Handler: newChatMember,
	}, &nats.Subscription{
		Subject: "core.product.flush",
		Handler: cache.FlushProduct,
	}, &nats.Subscription{
		Subject: "core.user.flush",
		Handler: cache.FlushUser,
	}, &nats.Subscription{
		Subject: "core.shop.flush",
		Handler: cache.FlushShop,
	})
}

func newMessage(req *chat.NewMessageRequest) {

	r := map[string]interface{}{
		"messages": req.Messages,
		"chat":     req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "retrieve", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
}

func messageReaded(req *chat.MessageReadedRequest) {

	r := map[string]interface{}{
		"message_id": req.MessageId,
		"user_id":    req.UserId,
		"chat":       req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "readed", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
}

func messageAppended(req *chat.MessageAppendedRequest) {

	r := map[string]interface{}{
		"message_id": req.Message.Id,
		"chat":       req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "appended", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
}

func newChatMember(req *chat.NewChatMemberRequest) {
	r := map[string]interface{}{
		"member": req.User,
		"chat":   req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("member", "joined", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
}
