package utils

import (
	"errors"
	"fmt"
	"proto/mail"
	"proto/sms"
	"utils/log"
	"utils/rpc"
	"core/api"
)

//EmailMessage is email message interface
type EmailMessage interface {
	GetFrom() string
	GetSubject() string
	GetMessage() string
	GetTo() string
}

//SmsMessage is sms message interface
type SmsMessage interface {
	GetMessage() string
	GetTo() string
}

// SendEmail sends email corresponding to provided fields
func SendEmail(msg EmailMessage) error {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	msgTo := SplitAndTrim(msg.GetTo())
	if len(msgTo) == 0 {
		return fmt.Errorf("Message has no recievers!")
	}

	request := &mail.MessageRequest{
		From:    msg.GetFrom(),
		Subject: msg.GetSubject(),
		Message: msg.GetMessage(),
		To:      msgTo,
	}

	log.Debug("Request: %#v", request)

	if msg.GetFrom() == "" || msg.GetSubject() == "" || msg.GetMessage() == "" {
		return fmt.Errorf("Email message can not have empty fields")
	}

	_, err := api.MailServiceClient.Send(ctx, request)

	return err
}

// SendSMS sends sms with this message fields
// fields.To should be a phone number
// fields.Subject should contain a message; the reason to store it here is qor which poisons .Body with html tags
// other fields are ignored
func SendSMS(msg SmsMessage) error {

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	//@TODO: check for number correctness?
	//@CHECK: is there a need for multiple recievers?

	if msg.GetTo() == "" || msg.GetMessage() == "" {
		return fmt.Errorf("Sms message should have both body (in subject field) and a reciever")
	}

	res, err := api.SmsServiceClient.SendSMS(ctx, &sms.SendSMSRequest{
		Phone: msg.GetTo(),
		Msg:   msg.GetMessage(),
	})

	if res.SmsError != "" {
		return errors.New(res.SmsError)
	}

	return err
}
