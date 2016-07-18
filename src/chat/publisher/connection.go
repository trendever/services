package publisher

import (
	"github.com/nats-io/nats"
	"utils/log"
	"chat/config"
	"chat/models"
)

const (
	EventJoin          = "chat.member.join"
	EventMessage       = "chat.message.new"
	EventMessageReaded = "chat.message.readed"
	//EventNotifySeller is a notification about a not answered message
	EventNotifySeller  = "core.notify.message"
	EventStatusMessage = "chat.status"
)

var cn *nats.Conn
var c *nats.EncodedConn
var chats models.ConversationRepository

func Init(repo models.ConversationRepository) {
	chats = repo
	conn, err := nats.Connect(config.Get().NatsURL)
	if err != nil {
		log.Fatal(err)
	}
	cn = conn
	c, err = nats.NewEncodedConn(cn, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}

	c.Subscribe(EventStatusMessage, addMessageStatus)
}

func Publish(subj string, data interface{}) {
	err := c.Publish(subj, data)
	if err != nil {
		log.Error(err)
	}
}
