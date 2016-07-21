package chat

import (
	"core/api"
	"encoding/json"
	"proto/chat"
	"utils/log"
	"utils/rpc"
)

//StatusContent is a representation of message content for chat status message
type StatusContent struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

//JSON returns json
func (cs StatusContent) JSON() string {
	content, err := json.Marshal(cs)
	log.Error(err)
	return string(content)
}

//SendChatMessage sends message to chat
func SendChatMessage(userID, conversationID uint64, content, mimeType string) error {
	context, cancel := rpc.DefaultContext()
	defer cancel()
	_, err := api.ChatServiceClient.SendNewMessage(context, &chat.SendMessageRequest{
		ConversationId: conversationID,
		Messages: []*chat.Message{
			{
				UserId: userID,
				Parts: []*chat.MessagePart{
					{
						Content:  string(content),
						MimeType: mimeType,
					},
				},
			},
		},
	})
	return err
}
