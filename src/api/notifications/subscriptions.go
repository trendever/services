package notifications

import (
	"api/cache"
	schat "api/chat"
	"common/log"
	"common/soso"
	"utils/nats"

	"proto/chat"
	"proto/core"
)

func init() {
	nats.StanSubscribe(
		// chat notifications
		&nats.StanSubscription{
			Subject:        "chat.message.new",
			DecodedHandler: newMessage,
		},
		&nats.StanSubscription{
			Subject:        "chat.message.readed",
			DecodedHandler: messageReaded,
		},
		&nats.StanSubscription{
			Subject:        "chat.message.appended",
			DecodedHandler: messageAppended,
		},
		&nats.StanSubscription{
			Subject:        "chat.member.join",
			DecodedHandler: newChatMember,
		},
		&nats.StanSubscription{
			Subject:        "chat.sync_event",
			DecodedHandler: syncEvent,
		},
		// core lead events
		&nats.StanSubscription{
			Subject:        "core.lead.event",
			DecodedHandler: onLeadEvent,
		},
	)
	nats.Subscribe(
		// core notifications
		&nats.Subscription{
			Subject: "core.product.flush",
			Handler: cache.FlushProduct,
		},
		&nats.Subscription{
			Subject: "core.user.flush",
			Handler: cache.FlushUser,
		},
		&nats.Subscription{
			Subject: "core.shop.flush",
			Handler: cache.FlushShop,
		},
	)
}

func syncEvent(chat *chat.Chat) bool {
	remoteCtx := soso.NewRemoteContext("chat", "sync_event", map[string]interface{}{
		"chat": chat,
	})
	schat.BroadcastMessage(chat.Members, nil, remoteCtx)
	return true
}

func newMessage(req *chat.NewMessageRequest) bool {

	r := map[string]interface{}{
		"messages": req.Messages,
		"chat":     req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "retrieve", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
	return true
}

func messageReaded(req *chat.MessageReadedRequest) bool {

	r := map[string]interface{}{
		"message_id": req.MessageId,
		"user_id":    req.UserId,
		"chat":       req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "readed", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
	return true
}

func messageAppended(req *chat.MessageAppendedRequest) bool {

	r := map[string]interface{}{
		"message_id": req.Message.Id,
		"chat":       req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("message", "appended", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
	return true
}

func newChatMember(req *chat.NewChatMemberRequest) bool {
	r := map[string]interface{}{
		"member": req.User,
		"chat":   req.Chat,
	}
	remoteCtx := soso.NewRemoteContext("member", "joined", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remoteCtx)
	return true
}

func onLeadEvent(req *core.LeadEventMessage) bool {

	log.Debug("Recieved lead event message for %v", req.Users)

	r := map[string]interface{}{
		"lead":           req.LeadId,
		"event":          req.Event,
		"users_can_join": len(req.Users),
	}

	ctx := soso.NewRemoteContext("lead", "event", r)
	broadcast(req.Users, ctx)
	return true
}

func broadcast(users []uint64, remoteCtx *soso.Context) {
	for _, userID := range users {
		sess := soso.Sessions.Get(userID)
		for _, ses := range sess {
			remoteCtx.Session = ses
			remoteCtx.SendResponse()
		}
	}
}
