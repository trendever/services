package server

import (
	"fmt"
	"golang.org/x/net/context"
	"proto/sms"
	"proto/telegram"
	"sms/conf"
	"sms/models"
	"utils/log"
	"utils/rpc"
)

type smsServer struct {
	sender        Sender
	telegramCli   telegram.TelegramServiceClient
	telegramRoom  string
	smsRepository models.SmsRepository
}

//Sender is interface for external service for sending sms
type Sender interface {
	SendSMS(*models.SmsDB) error
}

var senderFactories = map[string]func() (Sender, error){}

func RegisterSender(name string, factory func() (Sender, error)) {
	_, ok := senderFactories[name]
	if ok {
		log.Warn("Sender '%s' already registred", name)
	}
	senderFactories[name] = factory
}

func GetSender(name string) (Sender, error) {
	factory, ok := senderFactories[name]
	if !ok {
		return nil, fmt.Errorf("unknown sender '%v'", name)
	}
	sender, err := factory()
	if err != nil {
		return nil, fmt.Errorf("failed to create sender '%v': %v", name, err)
	}
	return sender, nil
}

//NewSmsServer returns new instance of *sms.SmsServiceServer
func NewSmsServer(sender Sender, smsRepository models.SmsRepository) sms.SmsServiceServer {
	s := conf.GetSettings().Telegram
	conn := rpc.Connect(s.RPC)
	return &smsServer{
		sender:        sender,
		telegramCli:   telegram.NewTelegramServiceClient(conn),
		telegramRoom:  s.Channel,
		smsRepository: smsRepository,
	}
}

//SendSMS sends sms
func (ss *smsServer) SendSMS(ctx context.Context, in *sms.SendSMSRequest) (*sms.SendSMSResult, error) {

	// new sms db object with status outgoing
	smsDbObj := &models.SmsDB{
		Phone:     in.Phone,
		Message:   in.Msg,
		SmsStatus: "outgoing",
	}

	// create new record
	if err := ss.smsRepository.Create(smsDbObj); err != nil {
		log.Error(err)
		return nil, err
	}

	go (func() {
		// send sms and return new sms db object
		if err := ss.sender.SendSMS(smsDbObj); err != nil {
			log.Error(err)
		}

		// save updated data to db
		if err := ss.smsRepository.Update(smsDbObj); err != nil {
			log.Error(err)
		}

		ctx, cancel := rpc.DefaultContext()
		defer cancel()
		_, err := ss.telegramCli.NotifyMessage(ctx, &telegram.NotifyMessageRequest{
			Channel: ss.telegramRoom,
			Message: fmt.Sprintf("SMS for %v with status '%v':\n%v", smsDbObj.Phone, smsDbObj.SmsStatus, smsDbObj.Message),
		})
		if err != nil {
			log.Errorf("failed to notify about new message: %v", err)
		}
	})()

	// return data through rpc
	smsResult := &sms.SendSMSResult{
		Id:        int64(smsDbObj.ID), // gorm's uint to int64
		SmsId:     smsDbObj.SmsID,
		SmsStatus: smsDbObj.SmsStatus,
		SmsError:  smsDbObj.SmsError,
	}

	return smsResult, nil
}

//RetrieveSmsStatus returns sms status
func (ss *smsServer) RetrieveSmsStatus(ctx context.Context, in *sms.RetrieveSmsStatusRequest) (*sms.RetrieveSmsStatusResult, error) {

	// new sms db object
	smsDbObj, err := ss.smsRepository.GetByID(uint(in.Id))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	// new sms status result object
	smsStatus := &sms.RetrieveSmsStatusResult{}

	// check if record with id is exist
	if smsDbObj == nil {
		// record not found
		smsStatus.Id = in.Id
		smsStatus.SmsStatus = "error"
		smsStatus.SmsError = "record not found"
	} else {
		// record found
		smsStatus.Id = in.Id
		smsStatus.SmsId = smsDbObj.SmsID
		smsStatus.SmsStatus = smsDbObj.SmsStatus
		smsStatus.SmsError = smsDbObj.SmsError
	}

	return smsStatus, nil
}
