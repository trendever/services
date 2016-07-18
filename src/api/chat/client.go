package chat

import (
	"proto/chat"
	"utils/rpc"
	"api/api"
	"api/soso"
)

var Client = chat.NewChatServiceClient(api.ChatConn)

func BroadcastMessage(members []*chat.Member, ctx, remote_ctx *soso.Context) {
	for _, member := range members {
		sess := soso.Sessions.Get(member.UserId)
		for _, ses := range sess {
			//skip current user only in current session
			if ctx != nil && ctx.Session.ID() == ses.ID() {
				continue
			}
			remote_ctx.Session = ses
			remote_ctx.SendResponse()
		}
	}
}

func GetChats(ids []uint64, userID uint64) ([]*chat.Chat, error) {
	ctx, cancell := rpc.DefaultContext()
	defer cancell()
	resp, err := Client.GetChats(ctx, &chat.ChatsRequest{Id: ids, UserId: userID})
	if err != nil {
		return nil, err
	}
	return resp.Chats, nil
}

func GetMessages(conversation_id, uid, from_message_id, limit uint64) (*chat.ChatHistoryReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	return Client.GetChatHistory(ctx, &chat.ChatHistoryRequest{
		ConversationId: conversation_id,
		UserId:         uid,
		FromMessageId:  from_message_id,
		Limit:          limit,
	})
}
