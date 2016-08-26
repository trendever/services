package pushers

import (
	"errors"
	"fmt"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"proto/push"
	"push/config"
	"time"
	"utils/log"
)

func init() {
	registerPusher(push.ServiceType_APN, &APNPusher{})
}

type APNPusher struct {
	cli   *apns2.Client
	topic string
}

var priorityMapAPN = map[push.Priority]int{
	push.Priority_NORMAL: apns2.PriorityLow,
	push.Priority_HING:   apns2.PriorityHigh,
}

func (p *APNPusher) Init() {
	config := config.Get()
	cert, err := certificate.FromPemFile(config.APNPemFile, config.APNPemPass)
	_ = err
	//if err != nil {
	//	log.Fatal(fmt.Errorf("failed to load APN certificate: %v", err))
	//}
	p.cli = apns2.NewClient(cert).Production()
	p.topic = config.APNTopic
}

func (s *APNPusher) Push(msg *push.PushMessage, tokens []string) (*PushResult, error) {
	priority, ok := priorityMapAPN[msg.Priority]
	if !ok {
		return nil, errors.New("unknown priority")
	}
	var ret PushResult
	for _, token := range tokens {
		res, err := s.cli.Push(&apns2.Notification{
			DeviceToken: token,
			Expiration:  time.Now().Add(time.Duration(msg.TimeToLive) * time.Second),
			Payload:     msg.Body,
			Priority:    priority,
		})
		// connection error
		if err != nil {
			ret.NeedRetry = append(ret.NeedRetry, token)
		}
		switch res.StatusCode {
		// success
		case 200:

		// 400: Bad request
		// 405: The request used a bad :method value. Only POST requests are supported
		// 413:	The notification payload was too large.
		case 400, 405, 413:
			return nil, fmt.Errorf("failed to send message: %v(%v)", res.Reason, res.StatusCode)

		// 410: The device token is no longer active for the topic
		case 410:
			ret.Invalids = append(ret.Invalids, token)

		// 429: The server received too many requests for the same device token.
		case 429:
			log.Warn("APNPusher: too many requests for token '%v', we will retry later")
			ret.NeedRetry = append(ret.NeedRetry, token)

		// 403: There was an error with the certificate
		// i think we need to save msg still for retry after reconfigure
		case 403:
			log.Error(errors.New("APNPusher: there was an error with the certificate"))
			ret.NeedRetry = append(ret.NeedRetry, token)

		// 500: Internal server error
		// 503: The server is shutting down and unavailable
		case 500, 503:
			ret.NeedRetry = append(ret.NeedRetry, token)

		// nothing else in documentation actuality
		default:
			log.Error(fmt.Errorf("APNPusher: unexpected HTTP status code %v", res.StatusCode))
			ret.NeedRetry = append(ret.NeedRetry, token)
		}
	}
	return &ret, nil
}
