package nats

import (
	"common/db"
	"common/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"reflect"
	"sync"
	"time"
)

type Config struct {
	URL         string
	StanCluster string
	StanID      string
}

type Subscription struct {
	Subject string
	Group   string
	Handler nats.Handler
}

type StanSubscription struct {
	Subject string
	Group   string
	// Set it make stan server remember about client after disconnection.
	// For non durable queue subscribers, when the last member leaves the group, that group is removed.
	// In individual subscription case durable name allows start receiving starting from last acknowledged by client massage
	DurableName string
	// if zero ack will be automatically called on msg receive
	AckTimeout time.Duration
	// normal stan handler, ack may be performed via msg.Ack()
	Handler stan.MsgHandler
	// easier way: data will be decoded with json.Unmarshal, ack will be performed after handler will return true,
	// should be func(decodedArg something) bool or func(decodedArg something, tx *gorm.DB) bool
	// in second case tx.Commit() will be called if handler will return true and msg.Ack() will be success, tx.Rollback() otherwise
	DecodedHandler interface{}
}

var (
	stanConn    stan.Conn
	encodedConn *nats.EncodedConn

	subscriptions     []*Subscription
	stanSubscriptions []*StanSubscription
	lock              sync.Mutex
)

// may be called before Init(), so it's fine to call it in package init
func Subscribe(subs ...*Subscription) {
	lock.Lock()
	for _, sub := range subs {
		subscriptions = append(subscriptions, sub)
		if encodedConn != nil {
			err := subscribe(sub)
			if err != nil {
				log.Errorf("failed to subscribe to NATS subject '%v': %v", sub.Subject, err)
			}
		}
	}
	lock.Unlock()
}

func subscribe(sub *Subscription) (err error) {
	if sub.Group == "" {
		_, err = encodedConn.Subscribe(sub.Subject, sub.Handler)
	} else {
		_, err = encodedConn.QueueSubscribe(sub.Subject, sub.Group, sub.Handler)
	}
	return err
}

func StanSubscribe(subs ...*StanSubscription) {
	lock.Lock()
	for _, sub := range subs {
		stanSubscriptions = append(stanSubscriptions, sub)
		if stanConn != nil {
			err := stanSubscribe(sub)
			if err != nil {
				log.Errorf("failed to subscribe to NATS Streaming subject '%v': %v", sub.Subject, err)
			}
		}
	}
	lock.Unlock()
}

func stanSubscribe(sub *StanSubscription) (err error) {
	if sub.DecodedHandler != nil {
		hType := reflect.TypeOf(sub.DecodedHandler)
		ok := true
		if hType.Kind() != reflect.Func {
			ok = false
		}
		if hType.NumOut() != 1 || hType.Out(0).Kind() != reflect.Bool {
			ok = false
		}
		hasTxArg := false
		switch {
		case hType.NumIn() == 1:
		case hType.NumIn() == 2:
			if hType.In(1) != reflect.TypeOf(db.New()) {
				ok = false
			} else {
				hasTxArg = true
			}
		default:
			ok = false
		}
		if !ok {
			return fmt.Errorf("DecodedHandler for subject %v has unexpected type", sub.Subject)
		}
		argType := hType.In(0)

		sub.Handler = func(m *stan.Msg) {
			var argPtr reflect.Value
			if argType.Kind() != reflect.Ptr {
				argPtr = reflect.New(argType)
			} else {
				argPtr = reflect.New(argType.Elem())
			}
			if err := json.Unmarshal(m.Data, argPtr.Interface()); err != nil {
				log.Errorf("failed to unmarshal argument to stan subscription %v: %v", sub.Subject, err)
				return
			}
			if argType.Kind() != reflect.Ptr {
				argPtr = reflect.Indirect(argPtr)
			}
			var args []reflect.Value
			var tx *gorm.DB
			if hasTxArg {
				tx = db.NewTransaction()
				args = []reflect.Value{argPtr, reflect.ValueOf(tx)}
			} else {
				args = []reflect.Value{argPtr}
			}
			hValue := reflect.ValueOf(sub.DecodedHandler)
			success := hValue.Call(args)[0].Bool()
			if !success {
				if hasTxArg {
					tx.Rollback()
				}
				return
			}
			if sub.AckTimeout != 0 {
				err := m.Ack()
				if err != nil {
					log.Errorf("failed to acknowledge nats server about successefuly handled msg: %v", err)
					if hasTxArg {
						tx.Rollback()
					}
					return
				}
			}
			if hasTxArg {
				tx.Commit()
			}
		}
	}

	var options = []stan.SubscriptionOption{}
	if sub.AckTimeout != 0 {
		options = append(options, stan.SetManualAckMode(), stan.AckWait(sub.AckTimeout))
	}
	if sub.DurableName != "" {
		options = append(options, stan.DurableName(sub.DurableName))
	}

	if sub.Group == "" {
		_, err = stanConn.Subscribe(sub.Subject, sub.Handler, options...)
	} else {
		_, err = stanConn.QueueSubscribe(sub.Subject, sub.Group, sub.Handler, options...)
	}
	return err
}

func Init(config *Config, stanRequired bool) {
	lock.Lock()
	defer lock.Unlock()
	for {
		err := connect(config, stanRequired)
		if err == nil {
			return
		}
		log.Error(err)
		time.Sleep(3 * time.Second)
	}
}

func connect(config *Config, stanRequired bool) error {
	conn, err := nats.Connect(config.URL, nats.MaxReconnects(-1))
	if err != nil {
		return fmt.Errorf("connection to NATS failed: %v", err)
	}
	encodedConn, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return fmt.Errorf("failed to create encoded NATS connection: %v", err)
	}
	for _, sub := range subscriptions {
		err := subscribe(sub)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to subscribe to NATS subject '%v': %v", sub.Subject, err)
		}
	}
	if config.StanCluster != "" && config.StanID != "" {
		stanConn, err = stan.Connect(config.StanCluster, config.StanID, stan.NatsConn(conn))
		if err != nil {
			return fmt.Errorf("connection to streaming server failed: %v", err)
		}
		for _, sub := range stanSubscriptions {
			err := stanSubscribe(sub)
			if err != nil {
				return fmt.Errorf("failed to subscribe to NATS Streaming subject '%v': %v", sub.Subject, err)
			}
		}
	} else if stanRequired {
		return errors.New("nats: stan reuqired but not configured")
	}
	return nil
}

func Publish(subj string, data interface{}) error {
	err := encodedConn.Publish(subj, data)
	if err != nil {
		log.Error(err)
	}
	return err
}

func StanPublish(subj string, data interface{}) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return stanConn.Publish(subj, encoded)
}
