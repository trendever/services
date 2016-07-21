package messager

import (
	"core/api"
	"core/conf"
	"github.com/nats-io/nats"
	"utils/log"
)

var (
	cn       *nats.Conn
	c        *nats.EncodedConn
	handlers = map[string]nats.Handler{}
)

//Init initializes nats connection
func Init() {
	conn, err := nats.Connect(conf.GetSettings().NatsURL)
	if err != nil {
		log.Fatal(err)
	}
	cn = conn
	c, err = nats.NewEncodedConn(cn, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}

	for subj, h := range handlers {
		c.Subscribe(subj, h)
	}

	go messageLoop()
}

//Publish publishes messages
func Publish(subj string, data interface{}) {
	err := c.Publish(subj, data)
	if err != nil {
		log.Error(err)
	}
}

func messageLoop() {
	for m := range api.Messages {
		Publish(m.Subj, m.Data)
	}
}
