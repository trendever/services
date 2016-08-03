package api

import (
	"payments/config"
	"payments/models"
	"proto/chat"
	"utils/rpc"
)

// Api connections
var (
	chatClient chat.ChatServiceClient
)

// ChatNotifier is mockable chat-notifier interface
type ChatNotifier interface {
	SendSessionToChat(sess *models.Session)
	SendPaymentToChat(pay *models.Payment)
}

type chatNotifierImpl struct {
	client chat.ChatServiceClient
}

// GetChatNotifier returns real ready-to-use chat notifier
func GetChatNotifier() ChatNotifier {
	return &chatNotifierImpl{
		client: chatClient,
	}
}

// Init initializes API connections
func Init() {
	settings := config.Get()
	chatClient = chat.NewChatServiceClient(rpc.Connect(settings.ChatServer))
}

// SendSessionToChat notifies chat about session update
func (cn *chatNotifierImpl) SendSessionToChat(sess *models.Session) error {

}

// SendPaymentToChat notifies chat about new payment order
func (cn *chatNotifierImpl) SendPaymentToChat(pay *models.Payment) error {

	message, err := json.Marshal(

	sendStatusMessage(
		sess.UserID,
		sess.ConversationId,
		string(message),
		"json/payment",
	)
}

func (cn *chatNotifierImpl) sendStatusMessage(userID, conversationID uint64, content, mimeType string) error {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := cn.client.SendNewMessage(context, &chat.SendMessageRequest{
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
