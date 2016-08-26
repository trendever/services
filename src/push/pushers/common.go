package pushers

import (
	"errors"
	"proto/push"
)

var registeredPushers = map[push.ServiceType]Pusher{}

type Update struct {
	Old string
	New string
}

type PushResult struct {
	// list of tokens that aren't valid
	Invalids []string
	// list of tokens for which push failed temporarily
	NeedRetry []string
	// list of tokens that should be updated
	Updates []Update
}

type Pusher interface {
	Init()
	Push(msg *push.PushMessage, tokens []string) (*PushResult, error)
}

func registerPusher(service push.ServiceType, pusher Pusher) {
	registeredPushers[service] = pusher
}

func Init() {
	for _, pusher := range registeredPushers {
		pusher.Init()
	}
}

func GetPusher(service push.ServiceType) (Pusher, error) {
	pusher, ok := registeredPushers[service]
	if !ok {
		return nil, errors.New("unknown push service type")
	}
	return pusher, nil
}
