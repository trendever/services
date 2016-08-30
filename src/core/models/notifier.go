package models

import (
	"core/api"
	"core/db"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"proto/mail"
	"proto/push"
	"proto/sms"
	"push/typemap"
	"reflect"
	"utils/log"
	"utils/rpc"
)

func init() {
	topics := []string{
		"notify_seller_about_lead",
		"notify_customer_about_lead",
		"notify_customer_about_unread_message",
		"notify_seller_about_unread_message",
		"call_supplier_to_chat",
		"call_customer_to_chat",
	}
	for _, t := range topics {
		RegisterTemplate("sms", t)
		RegisterTemplate("push", t)
		RegisterTemplate("email", t)
	}
}

// push notifications ttl in seconds
var PushTTL uint64 = 60 * 3

type Notifier struct {
	mailClient mail.MailServiceClient
	smsClient  sms.SmsServiceClient
	db         *gorm.DB
}

//NewNotifier returns new Notifier
func NewNotifier(mailClient mail.MailServiceClient, smsClient sms.SmsServiceClient, db *gorm.DB) *Notifier {
	return &Notifier{
		mailClient: mailClient,
		smsClient:  smsClient,
		db:         db,
	}
}

//GetNotifier returns initialized instance of Notifier
func GetNotifier() *Notifier {
	return NewNotifier(api.MailServiceClient, api.SmsServiceClient, db.New())
}

func (n *Notifier) NotifyByEmail(dest, about string, model interface{}) error {
	if dest == "" {
		return errors.New("destination address wasn't specified")
	}

	template := &EmailTemplate{}
	ret := n.db.Find(template, "template_id = ?", about)
	if ret.RecordNotFound() {
		log.Warn("Email template with ID '%v' not found", about)
		return nil
	}
	if ret.Error != nil {
		return ret.Error
	}
	result, err := template.Execute(model)
	if err != nil {
		return err
	}
	msg, ok := result.(*EmailMessage)
	if !ok {
		return errors.New("expected EmailMessage, but got " + reflect.TypeOf(msg).Name())
	}
	if msg.From == "" || msg.Subject == "" || msg.Body == "" {
		return errors.New("email message can not have empty fields")
	}

	request := &mail.MessageRequest{
		From:    msg.From,
		To:      []string{dest},
		Subject: msg.Subject,
		Message: msg.Body,
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	_, err = api.MailServiceClient.Send(ctx, request)

	return err
}

func (n *Notifier) NotifyBySms(phone, about string, model interface{}) error {
	if phone == "" {
		return errors.New("destination phone wasn't specified")
	}
	template := &SMSTemplate{}
	ret := n.db.Find(template, "template_id = ?", about)
	if ret.RecordNotFound() {
		log.Warn("SMS template with ID '%v' not found", about)
		return nil
	}
	if ret.Error != nil {
		return ret.Error
	}
	result, err := template.Execute(model)
	if err != nil {
		return err
	}
	msg, ok := result.(string)
	if !ok {
		return errors.New("expected string, but got " + reflect.TypeOf(msg).Name())
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	if msg == "" {
		return errors.New("empty message")
	}

	res, err := api.SmsServiceClient.SendSMS(ctx, &sms.SendSMSRequest{
		Phone: phone,
		Msg:   msg,
	})
	if err != nil {
		return err
	}

	if res.SmsError != "" {
		return errors.New(res.SmsError)
	}

	return err
}

func (n *Notifier) NotifyByPush(receivers []*push.Receiver, about string, model interface{}) error {
	if receivers == nil || len(receivers) == 0 {
		return errors.New("nil or empty receivers slice")
	}
	template := &PushTemplate{}
	ret := n.db.Find(template, "template_id = ?", about)
	if ret.RecordNotFound() {
		log.Warn("push template with ID '%v' not found", about)
		return nil
	}
	if ret.Error != nil {
		return ret.Error
	}
	result, err := template.Execute(model)
	if err != nil {
		return err
	}

	msg, ok := result.(*PushMessage)
	if !ok {
		return errors.New("expected PushMessage, but got " + reflect.TypeOf(msg).Name())
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	if msg.Data == "" && msg.Body == "" {
		return errors.New("empty message")
	}

	request := &push.PushRequest{
		Receivers: receivers,
		Message: &push.PushMessage{
			Priority:   push.Priority_HING,
			TimeToLive: PushTTL,
			Title:      msg.Title,
			Body:       msg.Body,
			Data:       msg.Data,
		},
	}
	_, err = api.PushServiceClient.Push(ctx, request)

	return err
}

// NotifyByTelegram sends string message to a channel
func (n *Notifier) NotifyByTelegram(channel string, message interface{}) error {
	return api.NotifyByTelegram(channel, message.(string))
}

// NotifyUserAbout sends to user notifications messages
// that are based on templates with TemplateId = about and execute argument context
func (n *Notifier) NotifyUserAbout(user *User, about string, context interface{}) error {
	log.Debug("Notify user %v about %v", user.Stringify(), about)
	var smsError, emailError error
	if user.Phone != "" {
		smsError = n.NotifyBySms(user.Phone, about, context)
	}
	if user.Email != "" {
		emailError = n.NotifyByEmail(user.Email, about, context)
	}

	var pushError error
	rep := GetPushTokensRepository()
	tokens, err := rep.GetTokens(user.ID)
	switch {
	case err != nil:
		pushError = err
	case tokens != nil && len(tokens) != 0:
		receivers := make([]*push.Receiver, 0, len(tokens))
		for _, token := range tokens {
			receivers = append(receivers, &push.Receiver{
				Service: typemap.TokenTypeToService[token.Type],
				Token:   token.Token,
			})
		}
		pushError = n.NotifyByPush(receivers, about, context)
	}

	if smsError == nil && emailError == nil && pushError == nil {
		return nil
	}
	return fmt.Errorf(
		"following errors happened while trying to notify user '%v' about %v:\n\tsms: %v\n\temail: %v\n\tpush: %v",
		user.Stringify(),
		about,
		smsError,
		emailError,
		pushError,
	)
}

func (n *Notifier) NotifySellerAboutLead(seller *User, lead *Lead) error {
	url, err := mkShortChatUrl(seller.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		seller,
		"notify_seller_about_lead",
		map[string]interface{}{
			"Seller": seller,
			"URL":    url,
			"Lead":   lead,
		},
	)
}

func (n *Notifier) NotifyCustomerAboutLead(customer *User, lead *Lead) error {
	url, err := mkShortChatUrl(customer.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		customer,
		"notify_customer_about_lead",
		map[string]interface{}{
			"Customer": customer,
			"URL":      url,
			"Lead":     lead,
		},
	)
}

func (n *Notifier) NotifySellerAboutUnreadMessage(seller *User, lead *Lead) error {
	url, err := mkShortChatUrl(seller.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		seller,
		"notify_seller_about_unread_message",
		map[string]interface{}{
			"Seller": seller,
			"URL":    url,
			"Lead":   lead,
		},
	)
}

func (n *Notifier) NotifyCustomerAboutUnreadMessage(customer *User, lead *Lead) error {
	url, err := mkShortChatUrl(customer.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		customer,
		"notify_customer_about_unread_message",
		map[string]interface{}{
			"Customer": customer,
			"URL":      url,
			"Lead":     lead,
		},
	)
}

func (n *Notifier) CallSupplierToChat(supplier *User, lead *Lead) error {
	url, err := mkShortChatUrl(supplier.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		supplier,
		"call_supplier_to_chat",
		map[string]interface{}{
			"Supplier": supplier,
			"URL":      url,
			"Lead":     lead,
		},
	)
}

func (n *Notifier) CallCustomerToChat(customer *User, lead *Lead) error {
	url, err := mkShortChatUrl(customer.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		customer,
		"call_customer_to_chat",
		map[string]interface{}{
			"Customer": customer,
			"URL":      url,
			"Lead":     lead,
		},
	)
}

func mkShortChatUrl(userId uint, leadId uint) (url string, err error) {
	// @CHECK Do we need long url with token? Why user authentication isn't enough?
	token, err := api.GetNewAPIToken(userId)
	if err != nil {
		return "", fmt.Errorf("can't get token for customer: %v", err)
	}
	url = api.GetChatURL(leadId, token)
	result, err := api.GetShortURL(url)
	if err != nil {
		// non-critical, we can return long url still
		log.Warn("GetShortURL: %v", err)
	} else {
		url = result.URL
	}
	return url, nil
}
