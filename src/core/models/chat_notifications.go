package models

import (
	"encoding/json"
	proto_chat "proto/chat"
	"core/api"
	"core/chat"
)

//SendProductToChat sends the product to the lead chat
func SendProductToChat(lead *Lead, product *Product) error {
	content, err := json.Marshal(product.Encode())
	if err != nil {
		return err
	}
	return chat.SendChatMessage(uint64(lead.CustomerID), lead.ConversationID, string(content), "text/json")
}

//SendStatusMessage sends status message
func SendStatusMessage(conversationID uint64, statusType, value string) {
	content := &chat.StatusContent{
		Type:  statusType,
		Value: value,
	}
	m := &proto_chat.Message{
		ConversationId: conversationID,
		Parts: []*proto_chat.MessagePart{
			&proto_chat.MessagePart{
				Content:  content.JSON(),
				MimeType: "json/status",
			},
		},
	}
	api.Publish("chat.status", m)
}
