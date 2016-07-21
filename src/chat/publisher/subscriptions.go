package publisher

import (
	"chat/models"
	"fmt"
	"proto/chat"
	"utils/log"
)

func addMessageStatus(m *chat.Message) {
	ch, err := chats.GetByID(uint(m.ConversationId))
	if err != nil {
		log.Error(err)
		return
	}

	if ch == nil {
		log.Error(fmt.Errorf("Unknow chat %v", m.ConversationId))
		return
	}

	msg := models.DecodeMessage(m, &models.Member{})
	if !msg.IsStatusMessage() {
		log.Error(fmt.Errorf("This subscriber accept only status messages!"))
		return
	}

	err = chats.AddMessages(ch, msg)
	if err != nil {
		log.Error(err)
		return
	}

	Publish(EventMessage, &chat.NewMessageRequest{
		Chat:     ch.Encode(),
		Messages: []*chat.Message{msg.Encode()},
	})
}
