package subscriber

import (
	"proto/chat"
	"proto/core"
	"api/cache"
	schat "api/chat"
	"api/soso"
)

func init() {
	//chat subscriptions
	handlers["chat.message.new"] = newMessage
	handlers["chat.message.readed"] = messageReaded
	handlers["chat.member.join"] = newChatMember
	handlers["core.product.new"] = newProduct
	handlers["core.product.update"] = updateProduct
}

func newMessage(req *chat.NewMessageRequest) {

	r := map[string]interface{}{
		"messages": req.Messages,
		"chat":     req.Chat,
	}
	remote_ctx := soso.NewRemoteContext("message", "retrieve", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remote_ctx)
}

func messageReaded(req *chat.MessageReadedRequest) {

	r := map[string]interface{}{
		"message_id": req.MessageId,
		"user_id":    req.UserId,
		"chat":       req.Chat,
	}
	remote_ctx := soso.NewRemoteContext("message", "readed", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remote_ctx)
}

func newChatMember(req *chat.NewChatMemberRequest) {
	r := map[string]interface{}{
		"member": req.User,
		"chat":   req.Chat,
	}
	remote_ctx := soso.NewRemoteContext("member", "joined", r)

	schat.BroadcastMessage(req.Chat.Members, nil, remote_ctx)
}

func updateProduct(product *core.Product) {
	cache.FlushProductCache(product.Id)
	for _, u := range product.LikedBy {
		switch {
		case u.InstagramUsername != "":
			cache.FlushUserCache(u.InstagramUsername)
		case u.Name != "":
			cache.FlushUserCache(u.Name)
		}
	}
}

func newProduct(product *core.Product) {

	cache.FlushProductCache(product.Id)

	resp := map[string]interface{}{
		"object_list": []interface{}{product},
		"count":       1,
	}

	remote_ctx := soso.NewRemoteContext("product", "new", resp)

	uSess := soso.Sessions.Get(uint64(product.MentionedId))
	sSess := soso.Sessions.Get(uint64(product.Supplier.SupplierId))

	for _, ses := range append(uSess, sSess...) {
		remote_ctx.Session = ses
		remote_ctx.SendResponse()
	}

}
