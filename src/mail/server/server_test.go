package server

import (
	"proto/mail"
	"golang.org/x/net/context"
	"mail/models"
	"testing"
)

type MailerStub struct {
}

type MailsStub struct {
}

func (m *MailerStub) Send(*models.Mail) error {
	return nil
}

func (m *MailsStub) Create(*models.Mail) error {
	return nil
}
func (m *MailsStub) Update(*models.Mail) error {
	return nil
}
func (m *MailsStub) GetByID(id uint) (*models.Mail, error) {
	return nil, nil
}

func TestSend(t *testing.T) {
	server := MakeNewMailServer(&MailerStub{}, &MailsStub{})
	_, err := server.Send(context.Background(), &mail.MessageRequest{
		From:    "some@email.com",
		Subject: "some subject",
		Message: "some message",
		To:      []string{"some@another.email"},
	})
	if err != nil {
		t.Errorf("Error duting send: %v", err)
	}
}

func TestStatus(t *testing.T) {
	server := MakeNewMailServer(&MailerStub{}, &MailsStub{})
	_, err := server.Status(context.Background(), &mail.StatusRequest{Id: 1})
	if err != nil {
		t.Errorf("Error duting status: %v", err)
	}
}
