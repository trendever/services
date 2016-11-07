package notifications

import (
	"fmt"

	"api/api"
	chatPkg "api/chat"
	"encoding/json"
	"proto/chat"
	"proto/payment"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

// Mime types
const (
	newOrder    = "json/order"
	cancelOrder = "json/cancel_order"
	newPayment  = "json/payment"
)

var paymentServiceClient = payment.NewPaymentServiceClient(api.PaymentsConn)

func init() {
	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "payments.event",
		Group:          "api",
		DecodedHandler: onPaymentEvent,
	})
}

func wrap(wr func(*payment.PaymentNotification) error, event *payment.PaymentNotification) bool {
	err := wr(event)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func updateData(id uint64, newData string) error {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := paymentServiceClient.UpdateServiceData(ctx, &payment.UpdateServiceDataRequest{
		Id:      id,
		NewData: newData,
	})

	return err
}

func decodeData(event *payment.PaymentNotification) (*payment.UsualData, error) {
	if event.Data == nil {
		return nil, fmt.Errorf("Wut? Got event with nil data")
	}

	// decode data
	var data payment.UsualData
	err := json.Unmarshal([]byte(event.Data.ServiceData), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func onPaymentEvent(event *payment.PaymentNotification) bool {

	switch event.Event {
	case payment.Event_Created:
		return wrap(sendPayment, event)
	case payment.Event_Cancelled:
		return wrap(sendCancelOrder, event)
	case payment.Event_PayFailed, payment.Event_PaySuccess:
		return wrap(sendSession, event)
	}

	return true
}

// SendSession notifies chat about session finish
func sendSession(event *payment.PaymentNotification) error {

	data, err := decodeData(event)
	if err != nil {
		return err
	}
	if data.MessageId == 0 {
		return fmt.Errorf("Zero message ID in payment(pay=%v); should not normally happen", event.Id)
	}

	var success = true
	if event.Event == payment.Event_PayFailed {
		success = false
	}

	// STEP1: send status message
	message, err := json.Marshal(&payment.ChatMessagePaymentFinished{
		PayId:     event.Id,
		Direction: data.Direction,

		Success: success,
		Failure: !success,

		Amount:   event.Data.Amount,
		Currency: event.Data.Currency,
	})

	if err != nil {
		return err
	}

	_, err = sendStatusMessage(
		event.Data.UserId,
		data.ConversationId,
		string(message),
		newPayment,
	)

	if err != nil {
		return err
	}

	err = appendStatusMessage(
		data.MessageId,
		string(message),
		newPayment,
	)

	if err != nil {
		return err
	}

	return nil
}

// SendPayment notifies chat about new payment order
func sendCancelOrder(event *payment.PaymentNotification) error {

	// Step1: notify chat about message
	message, err := json.Marshal(&payment.ChatMessageOrderCancelled{
		PayId:  uint64(event.Id),
		UserId: event.InvokerUserId,
	})
	if err != nil {
		return err
	}

	// decode data
	data, err := decodeData(event)
	if err != nil {
		return err
	}
	if data.MessageId == 0 {
		return fmt.Errorf("Zero MessageId in payment notification")
	}

	// append old message with upd info
	err = appendStatusMessage(
		data.MessageId,
		string(message),
		cancelOrder,
	)
	if err != nil {
		return err
	}

	// send new msg
	_, err = sendStatusMessage(
		event.InvokerUserId,
		data.ConversationId,
		string(message),
		cancelOrder,
	)

	if err != nil {
		return err
	}

	return nil
}

// SendPayment notifies chat about new payment order
func sendPayment(event *payment.PaymentNotification) error {

	data, err := decodeData(event)
	if err != nil {
		return err
	}

	// Step1: notify chat about message
	message, err := json.Marshal(&payment.ChatMessageNewOrder{
		PayId:    event.Id,
		Amount:   event.Data.Amount,
		Currency: event.Data.Currency,
	})

	if err != nil {
		return err
	}

	id, err := sendStatusMessage(
		event.Data.UserId,
		data.ConversationId,
		string(message),
		newOrder,
	)
	if err != nil {
		return err
	}

	// Step2: save message ID in payment
	data.MessageId = id
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return updateData(event.Id, string(encoded))
}

func sendStatusMessage(userID, conversationID uint64, content, mimeType string) (uint64, error) {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	res, err := chatPkg.Client.SendNewMessage(context, &chat.SendMessageRequest{
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

func appendStatusMessage(messageID uint64, content, mimeType string) error {
	context, cancel := rpc.DefaultContext()
	defer cancel()

	_, err := chatPkg.Client.AppendMessage(context, &chat.AppendMessageRequest{
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
