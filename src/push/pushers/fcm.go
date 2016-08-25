package pushers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/betrok/go-fcm"
	"proto/core"
	"proto/push"
	"push/config"
	"push/exteral"
	"utils/log"
	"utils/rpc"
)

func init() {
	registerPusher(push.ServiceType_FCM, NewFCMPusher(config.Get().FMCServerKey))
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

func NewFCMPusher(serverKey string) *FCMPusher {
	return &FCMPusher{serverKey: serverKey}
}

var priorityMapFCM = map[push.Priority]string{
	push.Priority_NORMAL: fcm.Priority_NORMAL,
	push.Priority_HING:   fcm.Priority_HIGH,
}

func (s *FCMPusher) Push(msg *push.PushMessage, tokens []string) (*PushResult, error) {
	cli := fcm.NewFcmClient(s.serverKey)
	cli.NewFcmRegIdsMsg(tokens, json.RawMessage(msg.Body))
	cli.SetTimeToLive(int(msg.TimeToLive))
	priority, ok := priorityMapFCM[msg.Prority]
	if !ok {
		return nil, errors.New("unknown priority")
	}
	cli.SetPriorety(priority)
	var ret PushResult
	res, err := cli.Send()
	// connection  error
	if err != nil {
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
		ret.NeedRetry = tokens
		return &ret, nil
	}
	for k, item := range res.Results {
		if err, ok := item["error"]; ok {
			switch err {
			case FCMMissingRegistration:
				// empty token?..
				continue

			case FCMInvalidRegistration, FCMNotRegistered, FCMInvalidPackageName, FCMMismatchSenderId:
				ret.Invalids = append(ret.Invalids, tokens[k])
				log.Error(fmt.Errorf("FCMPusher: token '%v' is invalid: %v", tokens[k], err))

			case FCMMessageTooBig, FCMInvalidDataKey, FCMInvalidTtl:
				log.Error(fmt.Errorf("FCMPusher: invalid message: %v", err))

			case FCMInternalServerError, FCMDeviceMessageRateExceeded, FCMTopicsMessageRateExceeded:
				log.Warn("FCMPusher: temporarily failed to send message for '%v': %v", tokens[k], err)
				ret.NeedRetry = append(ret.NeedRetry, tokens[k])

			}
		}
		if newToken, ok := item["registration_id"]; ok {
			go updateToken(tokens[k], newToken)
		}
	}
	return &ret, nil
}

func updateToken(oldToken, newToken string) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	exteral.PushTokensServiceClient.UpdateToken(ctx, &core.UpdateTokenRequest{
		Type:     core.TokenType_Android,
		OldToken: oldToken,
		NewToken: newToken,
	})
}
