package api

import (
	"encoding/json"
	"fmt"
	"payments/config"
	"payments/models"
	"proto/chat"
	"proto/payment"
	"utils/rpc"
)

// Api connections
var (
	chatClient chat.ChatServiceClient
)

// ChatNotifier is mockable chat-notifier interface
type ChatNotifier interface {
	// create pay button
	// return created message ID
	SendPaymentToChat(pay *models.Payment) error

	// use when session is in FINISHED state; create pay service msg; delete pay button
	SendSessionToChat(sess *models.Session) error
}

type chatNotifierImpl struct {
	client chat.ChatServiceClient
	repo   models.Repo
}

// GetChatNotifier returns real ready-to-use chat notifier
func GetChatNotifier(r models.Repo) ChatNotifier {
	return &chatNotifierImpl{
		client: chatClient,
		repo:   r,
	}
}

// Init initializes API connections
func Init() {
	settings := config.Get()
	chatClient = chat.NewChatServiceClient(rpc.Connect(settings.ChatServer))
}

// SendSessionToChat notifies chat about session finish
func (cn *chatNotifierImpl) SendSessionToChat(sess *models.Session) error {

	// STEP1: send status message
	message, err := json.Marshal(&payment.ChatMessagePaymentFinished{
		PayId:     uint64(sess.Payment.ID),
		Direction: payment.Direction(sess.Payment.Direction),

		Success: sess.Success,
		Amount:  sess.Amount,
	})

	if err != nil {
		return err
	}

	_, err = cn.sendStatusMessage(
		sess.Payment.UserID,
		sess.Payment.ConversationID,
		string(message),
		"json/payment",
	)

	if err != nil {
		return err
	}

	// STEP2: update old message with the same payment info
	if sess.Payment.MessageID == 0 {
		return fmt.Errorf("Zero message ID in payment(sess=%v,pay=%v); should not normally happen", sess.ID, sess.Payment.ID)
	}

	err = cn.appendStatusMessage(
		sess.Payment.MessageID,
		string(message),
		"json/payment",
	)

	if err != nil {
		return err
	}

	return nil
}

// SendPaymentToChat notifies chat about new payment order
func (cn *chatNotifierImpl) SendPaymentToChat(pay *models.Payment) error {

	// Step1: notify chat about message
	message, err := json.Marshal(&payment.ChatMessageNewOrder{
		PayId:    uint64(pay.ID),
		Amount:   pay.Amount,
		Currency: payment.Currency(pay.Currency),
	})

	if err != nil {
		return err
	}

	id, err := cn.sendStatusMessage(
		pay.UserID,
		pay.ConversationID,
		string(message),
		"json/order",
	)

	if err != nil {
		return err
	}

	// Step2: save message ID in payment
	pay.MessageID = id
	cn.repo.SavePay(pay)

	return nil
}

func (cn *chatNotifierImpl) sendStatusMessage(userID, conversationID uint64, content, mimeType string) (uint64, error) {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := cn.client.SendNewMessage(context, &chat.SendMessageRequest{
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

	if err != nil {
		return 0, err
	}

	if len(res.Messages) != 1 {
		return 0, fmt.Errorf("payments/chat: Wanted 1 message, but got %v; wut?", len(res.Messages))
	}

	return res.Messages[0].Id, nil
}

func (cn *chatNotifierImpl) appendStatusMessage(messageID uint64, content, mimeType string) error {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := cn.client.AppendMessage(context, &chat.AppendMessageRequest{
		MessageId: messageID,
		Parts: []*chat.MessagePart{
			{
				Content:  string(content),
				MimeType: mimeType,
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}
