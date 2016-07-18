package mailers

import (
	"github.com/mailgun/mailgun-go"
	"mail/models"
	"mail/server"
	"strings"
)

type Mailgun struct {
	api mailgun.Mailgun
}

func MakeNewMailgunMailer(domain, apiKey, publicApiKey string) server.Mailer {
	mg := mailgun.NewMailgun(domain, apiKey, publicApiKey)
	return &Mailgun{api: mg}
}

func (m *Mailgun) Send(email *models.Mail) error {
	msg := m.api.NewMessage(email.From, email.Subject, email.Message, strings.Split(email.To, ",")...)
	msg.SetHtml(email.Message)
	status, mid, err := m.api.Send(msg)
	email.Status = status
	email.Mid = mid
	return err
}
