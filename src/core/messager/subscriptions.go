package messager

import (
	"core/models"
	"proto/chat"
	"utils/log"
)

type chatRequest interface {
	GetChat() *chat.Chat
}

func init() {
	handlers[subscription{
		subject: "chat.message.new",
		group:   "core",
	}] = newMessage
	handlers[subscription{
		subject: "core.notify.message",
		group:   "core",
	}] = notifySellerAboutUnreadedMessage
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

	n := models.GetNotifier()
	for _, seller := range lead.Shop.Sellers {
		log.Error(n.NotifySellerAboutUnreadMessage(&seller, lead))
	}

	if lead.Shop.NotifySupplier {
		supplier, err := models.GetUserByID(lead.Shop.SupplierID)
		if err != nil {
			log.Error(err)
			return
		}
		log.Error(n.NotifySellerAboutUnreadMessage(supplier, lead))
	}
}
