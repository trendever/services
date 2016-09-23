package nats

import (
	"fmt"
	"github.com/nats-io/nats"
	"sync"
	"utils/log"
)

type Subscription struct {
	Subject string
	Group   string
	Handler nats.Handler
}

var (
	encoded       *nats.EncodedConn
	subscriptions []*Subscription
	lock          sync.Mutex
)

// may be called before Init(), so it's fine to call it in package init
func Subscribe(sub *Subscription) {
	lock.Lock()
	subscriptions = append(subscriptions, sub)
	if encoded != nil {
		err := subscribe(sub)
		if err != nil {
			log.Error(fmt.Errorf("failed to subscribe to NATS subject '%v': %v", sub.Subject, err))
		}
	}
	lock.Unlock()
}

func subscribe(sub *Subscription) (err error) {
	if sub.Group == "" {
		_, err = encoded.Subscribe(sub.Subject, sub.Handler)
	} else {
		_, err = encoded.QueueSubscribe(sub.Subject, sub.Group, sub.Handler)
	}
	return err
}

func Init(url string) {
	lock.Lock()
	defer lock.Unlock()
	conn, err := nats.Connect(url)
	if err != nil {
		log.Fatal(fmt.Errorf("connection to NATS failed: %v", err))
	}
	encoded, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create encoded NATS connection: %v", err))
	}

	for _, sub := range subscriptions {
		err := subscribe(sub)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to subscribe to NATS subject '%v': %v", sub.Subject, err))
		}
	}
}

func Publish(subj string, data interface{}) error {
	err := encoded.Publish(subj, data)
	if err != nil {
		log.Error(err)
	}
	return err
}
