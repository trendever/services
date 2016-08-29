package pushers

import (
	"encoding/json"
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

type APNAlert struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type APNPayload struct {
	Aps struct {
		Alert APNAlert `json:"alert,omitempty"`
	} `json:"aps,omitempty"`
	Data interface{} `json:"data,omitempty"`
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
	var payload APNPayload
	if msg.Data != "" {
		raw := json.RawMessage(msg.Data)
		payload.Data = &raw
	}
	if msg.Body != "" {
		payload.Aps.Alert = APNAlert{
			Title: msg.Title,
			Body:  msg.Body,
		}
	}
	for _, token := range tokens {
		res, err := s.cli.Push(&apns2.Notification{
			DeviceToken: token,
			Expiration:  time.Now().Add(time.Duration(msg.TimeToLive) * time.Second),
			Payload:     payload,
			Priority:    priority,
		})
		// connection error
		if err != nil {
			log.Debug("APNPusher: connection error: %v", err)
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
			log.Error(fmt.Errorf("APNPusher: token '%v' is invalid: %v", token, err))
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
			log.Warn("APNPusher: temporarily failed to send message for '%v': service unaviable(%v)", token, res.StatusCode)
			ret.NeedRetry = append(ret.NeedRetry, token)

		// nothing else in documentation actuality
		default:
			log.Error(fmt.Errorf("APNPusher: unexpected HTTP status code %v, reason: %v", res.StatusCode, res.Reason))
			ret.NeedRetry = append(ret.NeedRetry, token)
		}
	}
	return &ret, nil
}
