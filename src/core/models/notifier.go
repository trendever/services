package models

import (
	"core/api"
	"core/db"
	"core/notifier"
	"core/utils"
	"errors"
	"github.com/jinzhu/gorm"
	"proto/mail"
	"proto/sms"
	"reflect"
)

type notifierImpl struct {
	mailClient mail.MailServiceClient
	smsClient  sms.SmsServiceClient
	db         *gorm.DB
}

//NewNotifier returns new Notifier
func NewNotifier(mailClient mail.MailServiceClient, smsClient sms.SmsServiceClient, db *gorm.DB) notifier.Notifier {
	return &notifierImpl{
		mailClient: mailClient,
		smsClient:  smsClient,
		db:         db,
	}
}

//GetNotifier returns initialized instance of Notifier
func GetNotifier() notifier.Notifier {
	return NewNotifier(api.MailServiceClient, api.SmsServiceClient, db.New())
}

func (n *notifierImpl) NotifyByEmail(about string, model interface{}) error {
	template := &EmailTemplate{}
	if err := n.db.Find(template, "template_id = ?", about).Error; err != nil {
		return err
	}
	result, err := template.Parse(model)
	if err != nil {
		return err
	}
	msg, ok := result.(utils.EmailMessage)
	if !ok {
		return errors.New("Expected utils.EmailMessage, but got " + reflect.TypeOf(msg).Name())
	}
	return utils.SendEmail(msg)
}

func (n *notifierImpl) NotifyBySms(about string, model interface{}) error {
	template := &SMSTemplate{}
	if err := n.db.Find(template, "template_id = ?", about).Error; err != nil {
		return err
	}
	result, err := template.Parse(model)
	if err != nil {
		return err
	}
	msg, ok := result.(utils.SmsMessage)
	if !ok {
		return errors.New("Expected utils.SmsMessage, but got " + reflect.TypeOf(msg).Name())
	}
	return utils.SendSMS(msg)
}

// NotifyByTelegram sends string message to a channel
func (n *notifierImpl) NotifyByTelegram(channel string, message interface{}) error {
	return api.NotifyByTelegram(channel, message.(string))
}
