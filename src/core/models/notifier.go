package models

import (
	"core/api"
	"core/conf"
	"errors"
	"fmt"
	"proto/bot"
	"proto/chat"
	"proto/core"
	"proto/mail"
	"proto/push"
	"proto/sms"
	"proto/telegram"
	"push/typemap"
	"reflect"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"

	"github.com/jinzhu/gorm"
)

func init() {
	topics := []string{
		"notify_seller_about_lead",
		"notify_customer_about_lead",
		"notify_about_unanswered_message",
		"notify_user_about_new_messages",
		"call_supplier_to_chat",
		"call_customer_to_chat",
		"product_addad_for_seller",
		"product_addad_for_mentioner",
	}
	for _, t := range topics {
		RegisterNotifyTemplates(t)
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

	if msg == "" {
		return errors.New("empty message")
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

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

func (n *Notifier) NotifyUserByTelegram(user *User, about string, context interface{}) error {
	template := &TelegramTemplate{}
	ret := n.db.Find(template, "template_id = ?", about)
	if ret.RecordNotFound() {
		return nil
	}
	if ret.Error != nil {
		return ret.Error
	}
	result, err := template.Execute(context)
	if err != nil {
		return err
	}
	msg, ok := result.(string)
	if !ok {
		return errors.New("expected string, but got " + reflect.TypeOf(msg).Name())
	}
	if msg == "" {
		return nil
	}
	// may be not loaded yet
	if len(user.Telegram) == 0 {
		if err := db.New().Model(user).Related(&user.Telegram, "Telegram").Error; err != nil {
			return fmt.Errorf("failed to load related telegrams: %v", err)
		}
	}
	for _, tg := range user.Telegram {
		if !tg.Confirmed {
			continue
		}
		err := nats.StanPublish("telegram.notify", &telegram.NotifyMessageRequest{
			ChatId:  tg.ChatID,
			Message: msg,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// NotifyUserAbout sends to user notifications messages
// that are based on templates with TemplateId = about and execute argument context
func (n *Notifier) NotifyUserAbout(user *User, about string, context interface{}) error {
	log.Debug("Notify user %v about %v", user.Stringify(), about)
	var smsError, emailError, telgramError error
	if user.Phone != "" {
		smsError = n.NotifyBySms(user.Phone, about, context)
	}
	if user.Email != "" {
		emailError = n.NotifyByEmail(user.Email, about, context)
	}
	if user.HasTelegram {
		telgramError = n.NotifyUserByTelegram(user, about, context)
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

	if smsError == nil && emailError == nil && pushError == nil && telgramError == nil {
		return nil
	}
	strErr := fmt.Sprintf(
		"following errors happened while trying to notify user '%v' about %v:",
		user.Stringify(), about,
	)
	if smsError != nil {
		strErr += fmt.Sprintf("\n\t sms: %v", smsError)
	}
	if emailError != nil {
		strErr += fmt.Sprintf("\n\t email: %v", emailError)
	}
	if pushError != nil {
		strErr += fmt.Sprintf("\n\t push: %v", pushError)
	}
	if telgramError != nil {
		strErr += fmt.Sprintf("\n\t telegram: %v", telgramError)
	}
	return errors.New(strErr)
}

// loads user from db, appends him to context as 'user' and calls NotifyUserAbout method
func (n *Notifier) NotifyUserByID(userID uint64, about string, context map[string]interface{}) error {
	var user User
	err := db.New().First(&user, "id = ?", userID).Error
	if err != nil {
		return fmt.Errorf("failed to load user %v: %v", userID, err)
	}
	context["user"] = &user
	return n.NotifyUserAbout(&user, about, context)
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

func (n *Notifier) NotifyUserAboutNewMessages(user *User, lead *Lead, msgs []*chat.Message) error {
	url, err := mkShortChatUrl(user.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		user,
		"notify_user_about_new_messages",
		map[string]interface{}{
			"User":     user,
			"URL":      url,
			"Lead":     lead,
			"Messages": msgs,
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

func (n *Notifier) NotifyAboutUnansweredMessages(user *User, lead *Lead, count uint64, group string, msgs []*chat.Message) error {
	url, err := mkShortChatUrl(user.ID, lead.ID)
	if err != nil {
		return fmt.Errorf("failed to get lead url: %v", err)
	}
	return n.NotifyUserAbout(
		user,
		"notify_about_unanswered_message",
		map[string]interface{}{
			"User":     user,
			"URL":      url,
			"Lead":     lead,
			"Count":    count,
			"Group":    group,
			"Messages": msgs,
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

func (n *Notifier) NotifyAboutProductAdded(product *Product) {
	if product.MentionedBy.ID == 0 && product.MentionedByID > 0 {
		if user, err := GetUserByID(product.MentionedByID); err == nil {
			product.MentionedBy = *user
		}
	}

	if product.Shop.ID == 0 && product.ShopID > 0 {
		if shop, err := GetShopByID(product.ShopID, "Sellers"); err == nil {
			product.Shop = *shop
		}
	}

	url := fmt.Sprintf("%v/product/%v", conf.GetSettings().SiteURL, product.ID)
	log.Error(api.NotifyByTelegram(api.TelegramChannelNewProduct,
		fmt.Sprintf(
			"%v added %v by %v\n"+ // [scout] added [product_code] by [shop]
				"%v\n"+ // [instagram_link]
				"%v", // [qor_link]
			product.MentionedBy.Stringify(), product.Code, product.Shop.Stringify(),
			product.InstagramLink,
			fmt.Sprintf("%v/qor/products/%v", conf.GetSettings().SiteURL, product.ID),
		),
	))
	notify_map := map[uint]*User{}
	notify_map[product.Shop.SupplierID] = &product.Shop.Supplier
	for _, seller := range product.Shop.Sellers {
		notify_map[seller.ID] = seller
	}

	for _, user := range notify_map {
		log.Error(n.NotifyUserAbout(
			user,
			"product_addad_for_seller",
			map[string]interface{}{
				"user":    &product.MentionedBy,
				"product": product,
				"url":     url,
			},
		))
	}
	if _, ok := notify_map[product.MentionedByID]; !ok {
		log.Error(n.NotifyUserAbout(
			&product.MentionedBy,
			"product_addad_for_mentioner",
			map[string]interface{}{
				"user":    &product.MentionedBy,
				"product": product,
				"url":     url,
			},
		))
	}
	log.Error(nats.Publish("core.product.new", product.Encode()))

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

func SubmitCommentReply(lead *Lead) error {
	tmpl, err := GetOther(InstagramSubmitReplyTemplate)
	if err != nil {
		return err
	}

	res, err := tmpl.Execute(map[string]interface{}{
		"lead": lead,
	})
	if err != nil {
		return err
	}

	renderedString, ok := res.(string)
	if !ok || renderedString <= "" {
		return errors.New("String rendered to weird shit; skipping")
	}

	log.Debug("Requested to send `%v` to thread `%v`", renderedString, lead.InstagramMediaId)

	var req = bot.SendDirectRequest{
		SenderId: lead.Shop.Supplier.InstagramID,
		ThreadId: lead.InstagramMediaId,
		Type:     bot.MessageType_ReplyComment,
		ReplyKey: fmt.Sprintf("lead.%v.twat^Wsubmit", lead.ID), //change this when you need a reply %)
		Data:     renderedString,
	}
	err = nats.StanPublish("direct.send", &req)
	if err != nil {
		return fmt.Errorf("failed to send send comment request via nats: %v", err)
	}

	return nil
}

func mkShortChatUrl(userId uint, leadId uint) (url string, err error) {
	// @CHECK Do we need long url with token? Why user authentication isn't enough?
	token, err := api.GetNewAPIToken(userId)
	if err != nil {
		return "", fmt.Errorf("can't get token for customer: %v", err)
	}
	url = api.GetChatURL(leadId, token)

	return api.GetShortURL(url), nil
}

// NotifyUserCreated notifies about user creation
func NotifyUserCreated(u *User) {

	api.NotifyByTelegram(api.TelegramChannelNewUser,
		fmt.Sprintf(
			`#%v:
			New user %v registered
			%v`,
			u.Source,
			u.Stringify(),
			fmt.Sprintf("%v/qor/users/%v", conf.GetSettings().SiteURL, u.ID),
		),
	)
}

var actionText = map[core.LeadAction]string{
	core.LeadAction_BUY:  "ordered",
	core.LeadAction_INFO: "requested info about",
	core.LeadAction_SKIP: "skiped product",
}

// NotifyLeadCreated notifies about lead creation
func NotifyLeadCreated(l *Lead, p *Product, realInstLink string, action core.LeadAction) {

	if p.Shop.ID == 0 && p.ShopID > 0 {
		if shop, err := GetShopByID(p.ShopID); err == nil {
			p.Shop = *shop
		}
	}
	text := fmt.Sprintf(
		"%v %v %v by %v from %v, comment: '%v'\n%v\n", // [client] [action] [product_code] in [shop] from [wantit or website] comment: '[comment]' \n [qor_link]
		// first line
		l.Customer.Stringify(),
		actionText[action],
		p.Code,
		p.Shop.Stringify(),
		l.Source,
		l.Comment,
		fmt.Sprintf("%v/qor/orders/%v", conf.GetSettings().SiteURL, l.ID),
	)
	if l.IsNew() {
		text += "lead is new yet\n"
	} else {
		text += fmt.Sprintf("%v/chat/%v\n", conf.GetSettings().SiteURL, l.ID)
	}
	if realInstLink != "" {
		text += realInstLink + "\n"
	}
	// tag for search
	text += fmt.Sprintf("#%v", actionText[action])

	api.NotifyByTelegram(api.TelegramChannelNewLead, text)
}
