package messager

import (
	"core/api"
	"core/models"
	"core/notifier"
	"fmt"
	"proto/chat"
	"utils/log"
)

type chatRequest interface {
	GetChat() *chat.Chat
}

func init() {
	handlers["chat.message.new"] = newMessage
	handlers["core.notify.message"] = notifySellerAboutUnreadedMessage
}

func touchLead(req chatRequest) error {
	lead, err := models.GetLead(0, req.GetChat().Id)
	if err != nil {
		return err
	}
	return models.TouchLead(lead)
}

func newMessage(req *chat.NewMessageRequest) {
	log.Error(touchLead(req))
}

func notifySellerAboutUnreadedMessage(msg *chat.Message) {
	lead, err := models.GetLead(0, msg.ConversationId, "Shop", "Shop.Sellers", "Customer")
	if err != nil {
		log.Error(err)
		return
	}

	for _, seller := range lead.Shop.Sellers {
		notifySellerBySms(lead, seller)
	}

	if lead.Shop.NotifySupplier {
		supplier, err := models.GetUserByID(lead.Shop.SupplierID)
		if err != nil {
			log.Error(err)
			return
		}

		notifySellerBySms(lead, *supplier)
	}
}

func notifySellerBySms(lead *models.Lead, user models.User) {
	if user.Phone == "" {
		log.Error(fmt.Errorf("Seller or supplier without phone! [%v]%v", user.ID, user.GetName()))
		return
	}
	url, err := api.GetChatURLWithToken(lead.ID, user.ID)
	if err != nil {
		log.Error(err)
		return
	}
	r, err := api.GetShortURL(url)
	if err != nil {
		log.Error(err)
		return
	}
	err = notifier.NotifySellerAboutUnreadMessage(user, r.URL, lead, models.GetNotifier().NotifyBySms)
	log.Error(err)
}
