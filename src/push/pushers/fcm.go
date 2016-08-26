package pushers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/betrok/go-fcm"
	"proto/push"
	"push/config"
	"utils/log"
)

func init() {
	registerPusher(push.ServiceType_FCM, &FCMPusher{})
}

const (
	FCMMissingRegistration       = "MissingRegistration"
	FCMInvalidRegistration       = "InvalidRegistration"
	FCMNotRegistered             = "NotRegistered"
	FCMInvalidPackageName        = "InvalidPackageName"
	FCMMismatchSenderId          = "MismatchSenderId"
	FCMMessageTooBig             = "MessageTooBig"
	FCMInvalidDataKey            = "InvalidDataKey"
	FCMInvalidTtl                = "InvalidTtl"
	FCMInternalServerError       = "InternalServerError"
	FCMDeviceMessageRateExceeded = "DeviceMessageRateExceeded"
	FCMTopicsMessageRateExceeded = "TopicsMessageRateExceeded"
)

type FCMPusher struct {
	serverKey string
}

var priorityMapFCM = map[push.Priority]string{
	push.Priority_NORMAL: fcm.Priority_NORMAL,
	push.Priority_HING:   fcm.Priority_HIGH,
}

func (p *FCMPusher) Init() {
	p.serverKey = config.Get().FCMServerKey
}

func (s *FCMPusher) Push(msg *push.PushMessage, tokens []string) (*PushResult, error) {
	cli := fcm.NewFcmClient(s.serverKey)
	cli.AppendDevices(tokens)
	if msg.Data != "" {
		raw := json.RawMessage(msg.Data)
		cli.SetMsgData(&raw)
	}
	if msg.Body != "" {
		cli.SetNotificationPayload(
			&fcm.NotificationPayload{
				Title: msg.Title,
				Body:  msg.Body,
			},
		)
	}
	cli.SetTimeToLive(int(msg.TimeToLive))
	priority, ok := priorityMapFCM[msg.Priority]
	if !ok {
		return nil, errors.New("unknown priority")
	}
	cli.SetPriorety(priority)
	var ret PushResult
	res, err := cli.Send()
	// connection  error
	if err != nil {
		log.Debug("FCMPusher: connection error: %v, %+v", err, res)
		ret.NeedRetry = tokens
		return &ret, nil
	}
	switch res.StatusCode {
	case 200:

	case 400: // Bad request/Invalid JSON
		// probably invalid body in msg
		return nil, errors.New("invalid JSON")

	case 401: // Authentication Error
		log.Error(errors.New("FCMPusher: authentication error"))
		fallthrough
	default: // Unavailable
		log.Debug("service unaviable: %v", res.StatusCode)
		ret.NeedRetry = tokens
		return &ret, nil
	}
	for k, item := range res.Results {
		if err, ok := item["error"]; ok {
			switch err {
			case FCMMissingRegistration:
				log.Error(errors.New("FCMPusher: empty token provided"))
				continue

			case FCMInvalidRegistration, FCMNotRegistered, FCMInvalidPackageName, FCMMismatchSenderId:
				ret.Invalids = append(ret.Invalids, tokens[k])
				log.Error(fmt.Errorf("FCMPusher: token '%v' is invalid: %v", tokens[k], err))

			case FCMMessageTooBig, FCMInvalidDataKey, FCMInvalidTtl:
				log.Error(fmt.Errorf("FCMPusher: invalid message: %v", err))

			case FCMInternalServerError, FCMDeviceMessageRateExceeded, FCMTopicsMessageRateExceeded:
				log.Warn("FCMPusher: temporarily failed to send message for '%v': %v", tokens[k], err)
				ret.NeedRetry = append(ret.NeedRetry, tokens[k])

			default:
				log.Error(fmt.Errorf("FCMPusher: unknown error while sending to '%v': %v", tokens[k], err))
			}
		}
		if newToken, ok := item["registration_id"]; ok {
			ret.Updates = append(ret.Updates, Update{tokens[k], newToken})
		}
	}
	return &ret, nil
}
