package server

import (
	"golang.org/x/net/context"
	"proto/sms"
	"sms/models"
	"testing"
)

func TestSendSMS(t *testing.T) {
	s := NewSmsServer(&mockSender{}, &mockSmsRepository{})

	resp, err := s.SendSMS(context.Background(), &sms.SendSMSRequest{Phone: "some phone", Msg: "some message"})
	if err != nil {
		t.Errorf("Expected err must be nil, but got %q", err)
	}

	if resp.Id != 1 {
		t.Errorf("Expected ID must be 1, but got %q", resp.Id)
	}
}

func TestRetrieveSmsStatus(t *testing.T) {
	s := NewSmsServer(&mockSender{}, &mockSmsRepository{})

	resp, err := s.RetrieveSmsStatus(context.Background(), &sms.RetrieveSmsStatusRequest{Id: 1})

	if err != nil {
		t.Errorf("Expected err must be nil, but got %q", err)
	}

	if resp.Id != 1 {
		t.Errorf("Expected ID must be 1, but got %q", resp.Id)
	}
}

type mockSender struct {
}

func (ms *mockSender) SendSMS(*models.SmsDB) error {
	return nil
}

type mockSmsRepository struct {
}

func (ss *mockSmsRepository) Create(m *models.SmsDB) error {
	m.ID = 1
	return nil
}
func (ss *mockSmsRepository) Update(*models.SmsDB) error {
	return nil
}
func (ss *mockSmsRepository) GetByID(id uint) (*models.SmsDB, error) {
	m := &models.SmsDB{}
	m.ID = 1
	return m, nil
}
