package api

import (
	"fmt"
	"net/url"
	"core/conf"

	"github.com/timehop/go-bitly"
	"proto/auth"
	"proto/bot"
	"proto/chat"
	"proto/mail"
	"proto/sms"
	"utils/log"
	"utils/rpc"
	"google.golang.org/grpc"
)

type callbackFunc func(*grpc.Server)

//NatsMessage is nats message
type NatsMessage struct {
	Subj string
	Data interface{}
}

//Messages is chan for nats messages
var Messages = make(chan NatsMessage, 10)

var (
	server    *grpc.Server
	callbacks = make([]callbackFunc, 0)
)

// RPC Clients
var (
	MailServiceClient     mail.MailServiceClient
	SmsServiceClient      sms.SmsServiceClient
	ChatServiceClient     chat.ChatServiceClient
	AuthServiceClient     auth.AuthServiceClient
	TelegramServiceClient bot.TelegramServiceClient
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
	mailConn := rpc.Connect(conf.GetSettings().RPC.Mail)
	MailServiceClient = mail.NewMailServiceClient(mailConn)

	smsConn := rpc.Connect(conf.GetSettings().RPC.SMS)
	SmsServiceClient = sms.NewSmsServiceClient(smsConn)

	chatConn := rpc.Connect(conf.GetSettings().RPC.Chat)
	ChatServiceClient = chat.NewChatServiceClient(chatConn)

	authConn := rpc.Connect(conf.GetSettings().RPC.Auth)
	AuthServiceClient = auth.NewAuthServiceClient(authConn)

	telegramConn := rpc.Connect(conf.GetSettings().RPC.Telegram)
	TelegramServiceClient = bot.NewTelegramServiceClient(telegramConn)
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

//GetShortURL return short url for the url
func GetShortURL(url string) (bitly.ShortenResult, error) {
	return GetBitly().Shorten(url)
}

//GetChatURLWithToken returns url to chat with token
func GetChatURLWithToken(leadID uint, userID uint) (url string, err error) {
	token, err := GetNewAPIToken(userID)
	if err != nil {
		return
	}
	return GetChatURL(leadID, token), err
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
	if TelegramServiceClient != nil { // do nothing for tests and etc
		ctx, cancel := rpc.DefaultContext()
		defer cancel()

		_, err = TelegramServiceClient.NotifyMessage(ctx, &bot.NotifyMessageRequest{
			Channel: channel,
			Message: message,
		})

		if err != nil {
			log.Error(err)
		}
	}

	return err
}

//Publish sends message to nats
func Publish(subj string, data interface{}) {
	Messages <- NatsMessage{Subj: subj, Data: data}
}
