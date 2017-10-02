package server

import (
	"common/log"
	"golang.org/x/net/context"
	"mail/config"
	"mail/models"
	"proto/mail"
	"strings"
)

type Server struct {
	mailer             Mailer
	mails              models.MailRepository
	defaultFromAddress string
}

//Mailer is interface for external mail service
type Mailer interface {
	Send(*models.Mail) error
}

func MakeNewMailServer(mailer Mailer, mails models.MailRepository) mail.MailServiceServer {
	return &Server{mailer: mailer, mails: mails, defaultFromAddress: config.Get().From}
}

func (s *Server) Send(ctx context.Context, m *mail.MessageRequest) (*mail.StatusReply, error) {
	msg := &models.Mail{
		From:    m.From,
		Subject: m.Subject,
		Message: m.Message,
		To:      strings.Join(m.To, ","),
	}
	if msg.From == "" {
		msg.From = s.defaultFromAddress
	}
	if ok, err := msg.Validate(); !ok {
		return &mail.StatusReply{Id: uint64(msg.ID)}, err
	}
	if err := s.mails.Create(msg); err != nil {
		return &mail.StatusReply{Id: uint64(msg.ID)}, err
	}

	go (func(s *Server, msg *models.Mail) {
		if err := s.mailer.Send(msg); err != nil {
			log.Error(err)
			msg.StatusMsg = err.Error()
		}
		s.mails.Update(msg)

	})(s, msg)

	return &mail.StatusReply{Id: uint64(msg.ID)}, nil
}
func (s *Server) Status(ctx context.Context, m *mail.StatusRequest) (*mail.StatusReply, error) {
	msg, err := s.mails.GetByID(uint(m.Id))
	if msg == nil {
		return &mail.StatusReply{Id: 0, Status: "", Error: "Email not found"}, err
	}
	return &mail.StatusReply{Id: uint64(msg.ID), Status: msg.Status, Error: msg.StatusMsg}, err
}
