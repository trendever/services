package server

import (
	"golang.org/x/net/context"
	"proto/core"
	"proto/push"
	"push/config"
	"push/exteral"
	"push/models"
	"push/pushers"
	"push/typemap"
	"time"
	"utils/db"
	"utils/log"
	"utils/rpc"
)

const (
	MinimalRetryTimeout = 30
	RetryLoopTick       = 5
)

type PushServer struct {
	stop         chan struct{}
	retryTimeout uint64
}

func NewPushServer() *PushServer {
	ret := &PushServer{
		stop:         make(chan struct{}),
		retryTimeout: config.Get().RetryTimeout,
	}
	if ret.retryTimeout < MinimalRetryTimeout {
		log.Warn(
			"RetryTimeout %v is too small, use %v sec",
			ret.retryTimeout, MinimalRetryTimeout,
		)
		ret.retryTimeout = MinimalRetryTimeout
	}
	go ret.retryLoop()
	return ret
}

func (s *PushServer) Push(_ context.Context, in *push.PushRequest) (*push.PushResult, error) {
	notify := models.DecodeNotify(in)
	log.Debug("new push request: %+v", notify)
	retries := s.pushInternal(notify)
	// there is no need to save short-lived messages
	if in.Message.TimeToLive > config.Get().RetryTimeout/2 && len(retries) != 0 {
		go s.saveRetries(notify, retries)
	}
	return &push.PushResult{}, nil
}

func (s *PushServer) Stop() {
	close(s.stop)
}

// returns service -> []tokens map of receivers for which push failed temporarily
func (s *PushServer) pushInternal(notify *models.PushNotify) map[push.ServiceType][]string {
	tokens := notify.MapReceivers()
	retries := make(map[push.ServiceType][]string)
	for service, tokens := range tokens {
		pusher, err := pushers.GetPusher(service)
		if err != nil {
			log.Errorf("failed to get pusher %v: %v", service.String(), err)
			continue
		}
		res, err := pusher.Push(notify, tokens)
		if err != nil {
			log.Errorf("failed to send message via service %v: %v", service.String(), err)
			continue
		}
		if res.Invalids != nil {
			go invalidateTokens(service, res.Invalids)
		}
		if res.Updates != nil {
			go updateTokens(service, res.Updates)
		}
		if res.NeedRetry != nil {
			retries[service] = res.NeedRetry
		}
	}
	return retries
}

func (s *PushServer) saveRetries(notify *models.PushNotify, retries map[push.ServiceType][]string) {
	notify.LastTry = time.Now()
	notify.ReceiversFromMap(retries)
	if notify.Receivers == nil {
		return
	}
	err := db.New().Create(&notify).Error
	if err != nil {
		log.Errorf("failed to save retries: %v", err)
	}
}

func (s *PushServer) retryLoop() {
	for {
		select {
		case <-s.stop:
			return
		case <-time.Tick(time.Second * RetryLoopTick):
			now := time.Now()
			db.New().Delete(&models.PushNotify{}, "expiration < ?", now)
			db.New().Exec(`
				DELETE FROM push_receivers receiver
				WHERE NOT EXISTS (
					SELECT 1 FROM push_notifies_receivers relation
					WHERE relation.push_receiver_id = receiver.id
				)
			`)
			var saved []models.PushNotify
			err := db.New().Preload("Receivers").Limit(100).Find(&saved, "last_try < ?", now.Truncate(time.Duration(s.retryTimeout)*time.Second)).Error
			if err != nil {
				log.Errorf("failed to load reties: %v", err)
				continue
			}
			if len(saved) == 0 {
				continue
			}
			for _, notify := range saved {
				retries := s.pushInternal(&notify)
				notify.ReceiversFromMap(retries)
				if notify.Receivers == nil {
					err := db.New().Delete(&notify).Error
					if err != nil {
						log.Errorf("failed to delete resended notify: %v", err)
					}
				} else {
					notify.LastTry = now
					err := db.New().Find(&notify.Receivers).Error
					if err != nil {
						log.Errorf("failed to save updated notify: %v", err)
					}
					err = db.New().Save(&notify).Error
					if err != nil {
						log.Errorf("failed to save updated notify: %v", err)
					}
				}
			}
		}
	}
}

func invalidateTokens(service push.ServiceType, tokens []string) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	exteral.PushTokensServiceClient.InvalidateTokens(ctx, &core.InvalidateTokensRequest{
		Type:   typemap.ServiceToTokenType[service],
		Tokens: tokens,
	})
}

func updateTokens(service push.ServiceType, updates []pushers.Update) {
	tokenType := typemap.ServiceToTokenType[service]
	for _, up := range updates {
		ctx, cancel := rpc.DefaultContext()
		exteral.PushTokensServiceClient.UpdateToken(ctx, &core.UpdateTokenRequest{
			Type:     tokenType,
			OldToken: up.Old,
			NewToken: up.New,
		})
		cancel()
	}
}
