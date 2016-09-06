package views

import (
	"api/chat"
	"api/soso"
	"errors"
	"golang.org/x/net/context"
	"net/http"
	proto_chat "proto/chat"
	"proto/core"
	"time"
	"utils/rpc"
)

var chatClient = chat.Client

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"create", "message", ChatSendMessage},
		soso.Route{"update", "message", ChatMessageMarkAsRead},
		soso.Route{"count_unread", "message", ChatTotalCountUnread},
		soso.Route{"join", "chat", ChatJoin},
		soso.Route{"leave", "chat", ChatLeave},
		soso.Route{"search", "message", ChatHistory},
		soso.Route{"list", "chat", ChatList},
		soso.Route{"call_supplier", "chat", CallSupplierToChat},
		soso.Route{"call_customer", "chat", CallCustomerToChat},
	)
}

func ChatSendMessage(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	part := &proto_chat.MessagePart{
		MimeType: "text/plain",
	}
	message := &proto_chat.Message{
		UserId: c.Token.UID,
		Parts: []*proto_chat.MessagePart{
			part,
		},
	}
	request := &proto_chat.SendMessageRequest{
		Messages: []*proto_chat.Message{
			message,
		},
	}

	if value, ok := req["text"].(string); ok {
		part.Content = value
	}

	if value, ok := req["mime_type"].(string); ok && value != "" {
		part.MimeType = value
	}

	if value, ok := req["conversation_id"].(float64); ok {
		request.ConversationId = uint64(value)
	}

	if part.Content == "" || request.ConversationId == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Text and conversation_id are required"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	resp, err := chatClient.SendNewMessage(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}
	if resp.Error != nil {
		c.Response.ResponseMap = map[string]interface{}{
			"error":    resp.Error,
			"messages": resp.Messages,
			"chat":     resp.Chat,
		}
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New(resp.Error.Message))
		return
	}
	//r := map[string]interface{}{
	//	"messages": resp.Messages,
	//	"chat":     resp.Chat,
	//}
	//remote_ctx := soso.NewRemoteContext("message", "retrieve", r)
	//
	//go chat.BroadcastMessage(resp.Chat.Members, c, remote_ctx)

	//conversation id, message
	c.SuccessResponse(map[string]interface{}{
		"error":    resp.Error,
		"messages": resp.Messages,
		"chat":     resp.Chat,
	})
}

func ChatJoin(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	user, err := GetUser(c.Token.UID, false)

	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	req := c.RequestMap
	member := &proto_chat.Member{
		UserId: c.Token.UID,
		Name:   user.GetName(),
	}
	request := &proto_chat.JoinChatRequest{
		Members: []*proto_chat.Member{
			member,
		},
	}
	var lead_id, conversation_id uint64
	if value, ok := req["lead_id"].(float64); ok {
		lead_id = uint64(value)
	}

	if value, ok := req["conversation_id"].(float64); ok {
		conversation_id = uint64(value)
	}

	if lead_id == 0 && conversation_id == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("lead_id or conversation_id are required"))
		return
	}

	//detecting user role
	{
		leadRequest := &core.GetLeadRequest{
			UserId: c.Token.UID,
		}

		switch {
		case lead_id > 0:
			leadRequest.SearchBy = &core.GetLeadRequest_Id{Id: lead_id}
		case conversation_id > 0:
			leadRequest.SearchBy = &core.GetLeadRequest_ConversationId{ConversationId: conversation_id}
		}

		ctx, cancel := rpc.DefaultContext()
		defer cancel()
		resp, err := leadServiceClient.GetLead(ctx, leadRequest)
		if err != nil {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
			return
		}
		if resp.Lead == nil {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Can't find lead"))
			return
		}
		lead := resp.Lead
		if lead.ConversationId == 0 {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Oops, lead without chat. Try to change state of lead to NEW"))
			return
		}
		//convert LeadUserRole to MemberRole
		role, ok := proto_chat.MemberRole_value[lead.UserRole.String()]
		if !ok {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("You can't join to this chat"))
			return
		}
		member.Role = proto_chat.MemberRole(role)
		request.ConversationId = lead.ConversationId

	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := chatClient.JoinChat(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}
	//conversation id
	c.SuccessResponse(map[string]interface{}{
		"error": resp.Error,
		"chat":  resp.Chat,
	})
}

func ChatLeave(c *soso.Context) {
	//conversation id
	c.SuccessResponse(map[string]interface{}{
		"status": "not implemented",
	})
}

func ChatHistory(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &proto_chat.ChatHistoryRequest{
		UserId: c.Token.UID,
	}

	if value, ok := req["conversation_id"].(float64); ok {
		request.ConversationId = uint64(value)
	}

	if value, ok := req["limit"].(float64); ok {
		request.Limit = uint64(value)
	}

	if value, ok := req["from_message_id"].(float64); ok {
		request.FromMessageId = uint64(value)
	}

	if value, ok := req["direction"].(bool); ok {
		request.Direction = value
	}

	if request.ConversationId == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Conversation_id is required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := chatClient.GetChatHistory(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	//conversation id, date_from
	c.SuccessResponse(map[string]interface{}{
		"error":    resp.Error,
		"messages": resp.Messages,
	})
}

func ChatList(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	request := &proto_chat.UserChatsRequest{
		UserId: c.Token.UID,
	}

	ctx, cancell := rpc.DefaultContext()
	defer cancell()
	resp, err := chatClient.GetUserChats(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"error": resp.Error,
		"chats": resp.Chats,
	})
}

func CallSupplierToChat(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &core.CallSupplierRequest{}

	if value, ok := req["lead_id"].(float64); ok {
		request.LeadId = uint64(value)
	}

	ctx, cancell := rpc.DefaultContext()
	defer cancell()
	_, err := leadServiceClient.CallSupplier(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "ok",
	})
}

func CallCustomerToChat(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &core.CallCustomerRequest{}

	if value, ok := req["lead_id"].(float64); ok {
		request.LeadId = uint64(value)
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	_, err := leadServiceClient.CallCustomer(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "ok",
	})
}

func ChatMessageMarkAsRead(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	req := c.RequestMap
	request := &proto_chat.MarkAsReadedRequest{
		UserId: c.Token.UID,
	}

	if value, ok := req["conversation_id"].(float64); ok {
		request.ConversationId = uint64(value)
	}

	if value, ok := req["message_id"].(float64); ok {
		request.MessageId = uint64(value)
	}

	if request.MessageId == 0 || request.ConversationId == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("conversation_id and message_id is required"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := chatClient.MarkAsReaded(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	if resp.Error != nil {
		c.Response.ResponseMap = map[string]interface{}{
			"error": resp.Error,
		}
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New(resp.Error.Message))
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"status": "ok",
	})

}

func getChats(chIDs []uint64, uid uint64) (chats []*proto_chat.Chat, err error) {
	if len(chIDs) > 0 {
		chats, err = chat.GetChats(chIDs, uid)
	}
	return chats, err
}

func getChat(id, uid uint64) (*proto_chat.Chat, error) {
	chats, err := getChats([]uint64{id}, uid)
	if err == nil && len(chats) == 1 {
		return chats[0], nil
	}
	return nil, err
}

func ChatTotalCountUnread(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}
	request := &proto_chat.TotalCountUnreadRequest{
		UserId: c.Token.UID,
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	resp, err := chatClient.GetTotalCountUnread(ctx, request)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"count": resp.Count,
	})
}
