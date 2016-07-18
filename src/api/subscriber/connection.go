package subscriber

import (
	"github.com/nats-io/nats"
	"utils/log"
	"api/conf"
)

var (
	cn       *nats.Conn
	c        *nats.EncodedConn
	handlers = map[string]nats.Handler{}
)

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
}
