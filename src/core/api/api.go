package api

import (
	"core/conf"
	"fmt"
	"net/url"
	"proto/auth"
	"proto/chat"
	"proto/checker"
	"proto/mail"
	"proto/payment"
	"proto/push"
	"proto/sms"
	"proto/telegram"
	"proto/trendcoin"
	"utils/log"
	"utils/nats"
	"utils/rpc"

	bitly "github.com/timehop/go-bitly"
	"google.golang.org/grpc"
)

type callbackFunc func(*grpc.Server)

var (
	server    *grpc.Server
	callbacks = make([]callbackFunc, 0)
)

// RPC Clients
var (
	MailServiceClient      mail.MailServiceClient
	SmsServiceClient       sms.SmsServiceClient
	ChatServiceClient      chat.ChatServiceClient
	AuthServiceClient      auth.AuthServiceClient
	PushServiceClient      push.PushServiceClient
	CheckerServiceClient   checker.CheckerServiceClient
	TrendcoinServiceClient trendcoin.TrendcoinServiceClient
	PaymentsServiceClient  payment.PaymentServiceClient
)

// Telegram channel destanations
const (
	TelegramChannelNewUser    = "new_user"
	TelegramChannelNewLead    = "new_lead"
	TelegramChannelNewProduct = "new_product"
)

// Start initializes server listening
func Start() {
	server = rpc.Serve(conf.GetSettings().RPC.Listen)

	// callbacks are used to attach core service implementations
	// see service/views/
	for _, cb := range callbacks {
		cb(server)
	}

	startClients()
}

// AddOnStartCallback executed when server is initialized
func AddOnStartCallback(cb callbackFunc) {
	callbacks = append(callbacks, cb)
}

// startClients starts RPC connections to external services
func startClients() {
	config := conf.GetSettings()
	mailConn := rpc.Connect(config.RPC.Mail)
	MailServiceClient = mail.NewMailServiceClient(mailConn)

	smsConn := rpc.Connect(config.RPC.SMS)
	SmsServiceClient = sms.NewSmsServiceClient(smsConn)

	chatConn := rpc.Connect(config.RPC.Chat)
	ChatServiceClient = chat.NewChatServiceClient(chatConn)

	authConn := rpc.Connect(config.RPC.Auth)
	AuthServiceClient = auth.NewAuthServiceClient(authConn)

	pushConn := rpc.Connect(config.RPC.Push)
	PushServiceClient = push.NewPushServiceClient(pushConn)

	checkerConn := rpc.Connect(config.RPC.Checker)
	CheckerServiceClient = checker.NewCheckerServiceClient(checkerConn)

	trendcoinConn := rpc.Connect(config.RPC.Trendcoin)
	TrendcoinServiceClient = trendcoin.NewTrendcoinServiceClient(trendcoinConn)

	paymentsConn := rpc.Connect(config.RPC.Payments)
	PaymentsServiceClient = payment.NewPaymentServiceClient(paymentsConn)
}

// GetBitly returns Bitly client
func GetBitly() *bitly.Client {
	settings := conf.GetSettings().Bitly
	return &bitly.Client{APIKey: settings.APIKey, Login: settings.Login, AccessToken: settings.AccessToken}
}

//GetChatURL returns url to chat
func GetChatURL(leadID uint, token string) string {
	v := &url.Values{}
	v.Add("token", token)
	u, err := url.Parse(conf.GetSettings().SiteURL)
	if err != nil {
		log.Error(err)
		return ""
	}
	u.Path = fmt.Sprintf("chat/%v", leadID)
	u.RawQuery = v.Encode()
	return u.String()
}

//GetMonetizationURL returns url to monetization
func GetMonetizationURL(userID uint64) string {
	return AddUserToken(conf.GetSettings().URL.Monetization, userID)
}

func AddUserToken(link string, userID uint64) string {
	token, err := GetNewAPIToken(uint(userID))
	if err != nil {
		log.Errorf("failed to get token for user %v: %v", userID, err)
		return ""
	}
	parsed, err := url.Parse(link)
	if err != nil {
		log.Errorf("invalid URL passed to AddToken: \"%v\", %v", link, err)
		return ""
	}
	values := parsed.Query()
	values.Set("token", token)
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

//GetShortURL return short url for the url
func GetShortURL(url string) string {
	short, err := GetBitly().Shorten(url)
	if err != nil {
		// non-critical, we can return long url still
		log.Warn("GetShortURL: %v", err)
		return url
	}
	return short.URL
}

// GetNewAPIToken returns login API token
func GetNewAPIToken(userID uint) (token string, err error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := AuthServiceClient.GetNewToken(ctx, &auth.NewTokenRequest{UserId: uint64(userID)})
	if err != nil {
		return
	}
	token = resp.Token
	return
}

// NotifyByTelegram sends string message to a channel
func NotifyByTelegram(channel, message string) (err error) {
	err = nats.StanPublish("telegram.notify", &telegram.NotifyMessageRequest{
		Channel: channel,
		Message: message,
	})

	if err != nil {
		log.Error(err)
	}

	return err
}
