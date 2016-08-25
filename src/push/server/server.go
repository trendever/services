package server

import (
	"fmt"
	"golang.org/x/net/context"
	"proto/core"
	"proto/push"
	"push/exteral"
	"push/pushers"
	"push/typemap"
	"utils/log"
	"utils/rpc"
)

type PushServer struct {
	stop chan struct{}
}

func NewPushServer() *PushServer {
	return &PushServer{
		stop: make(chan struct{}),
	}
}

func (s *PushServer) Push(_ context.Context, in *push.PushRequest) (*push.PushResult, error) {
	tokens := map[push.ServiceType][]string{}
	for _, receiver := range in.Receivers {
		tokens[receiver.Service] = append(tokens[receiver.Service], receiver.Token)
	}
	for service, tokens := range tokens {
		pusher, err := pushers.GetPusher(service)
		if err != nil {
			log.Error(fmt.Errorf("failed to get pusher %v: %v", service.String(), err))
			continue
		}
		res, err := pusher.Push(in.Message, tokens)
		if err != nil {
			log.Error(fmt.Errorf("failed to send message via service %v: %v", service.String(), err))
			continue
		}
		if res.Invalids != nil {
			go invalidateTokens(service, res.Invalids)
		}
	}
	return &push.PushResult{}, nil
}

func (s *PushServer) Stop() {
	close(s.stop)
}

func invalidateTokens(service push.ServiceType, tokens []string) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	exteral.PushTokensServiceClient.InvalidateTokens(ctx, &core.InvalidateTokensRequest{
		Type:  typemap.ServiceToTokenType[service],
		Tokens: tokens,
	})
}
