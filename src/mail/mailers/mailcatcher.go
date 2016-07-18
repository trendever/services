package mailers

import (
	"net/smtp"
	"mail/models"
	"mail/server"
	"strings"
)

type mailcatcher struct {
	Addr string
}

func NewMailcatcher(addr string) server.Mailer {
	return &mailcatcher{Addr: addr}
}

func (m mailcatcher) Send(model *models.Mail) error {
	body := []byte(
		"Subject: " + model.Subject + "\r\n" +
			"Content-Type: text/html \r\n" +
			"\r\n" +
			model.Message,
	)
	return smtp.SendMail(m.Addr, nil, model.From, strings.Split(model.To, ","), body)
}
