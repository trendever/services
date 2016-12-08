package senders

import (
	"sms/models"
	"sms/server"
)

func init() {
	server.RegisterSender("dummy", NewDummySender)
}

type dummySender struct{}

func NewDummySender() (server.Sender, error) {
	return &dummySender{}, nil
}

func (s *dummySender) SendSMS(sms *models.SmsDB) error {
	sms.SmsStatus = "sent"
	return nil
}
