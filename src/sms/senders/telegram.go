package senders

import (
	"fmt"
	"proto/bot"
	"sms/conf"
	"sms/models"
	"sms/server"
	"utils/rpc"
)

func init() {
	server.RegisterSender("telegram", NewTelesender)
}

type Telesender struct {
	client  bot.TelegramServiceClient
	channel string
}

func NewTelesender() (server.Sender, error) {
	s := conf.GetSettings().Telegram
	conn := rpc.Connect(s.Rpc)
	return &Telesender{
		client:  bot.NewTelegramServiceClient(conn),
		channel: s.Channel,
	}, nil
}

func (s *Telesender) SendSMS(sms *models.SmsDB) error {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	_, err := s.client.NotifyMessage(ctx, &bot.NotifyMessageRequest{
		Channel: s.channel,
		Message: fmt.Sprintf("New SMS for %v: %v", sms.Phone, sms.Message),
	})
	if err != nil {
		sms.SmsStatus = "failed"
		sms.SmsError = err.Error()
	} else {
		sms.SmsStatus = "sent"
	}
	return nil
}
