package subscriber

import (
	"api/cache"
	schat "api/chat"
	"api/soso"
	"proto/chat"
)

func init() {
	//chat subscriptions
	handlers["chat.message.new"] = newMessage
	handlers["chat.message.readed"] = messageReaded
	handlers["chat.member.join"] = newChatMember

	handlers["core.product.flush"] = cache.FlushProduct
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

func newChatMember(req *chat.NewChatMemberRequest) {
	r := map[string]interface{}{
		"member": req.User,
		"chat":   req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("member", "joined", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
}
