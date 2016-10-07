package views

import (
	"core/models"
	"fmt"
	"proto/chat"
	"proto/core"
	"time"
	"utils/db"
	"utils/log"
	"utils/nats"
)

func init() {
	nats.Subscribe(&nats.Subscription{
		Subject: "chat.message.new",
		Group:   "core",
		Handler: newMessage,
	})
	nats.Subscribe(&nats.Subscription{
		Subject: "core.notify.message",
		Group:   "core",
		Handler: notifySellerAboutUnreadedMessage,
	})
	nats.Subscribe(&nats.Subscription{
		Subject: "auth.login",
		Group:   "core",
		Handler: handleUserLogin,
	})
	nats.Subscribe(&nats.Subscription{
		Subject: "api.new_session",
		Group:   "core",
		Handler: handleNewSession,
	})
}

func newMessage(req *chat.NewMessageRequest) {
	log.Error(models.TouchLead(req.Chat.Id))
}

func notifySellerAboutUnreadedMessage(msg *chat.Message) {
	lead, err := models.GetLead(0, msg.ConversationId, "Shop", "Shop.Sellers", "Customer")
	if err != nil {
		log.Error(err)
		return
	}

	if lead.State == core.LeadStatus_NEW.String() {
		return
	}

	n := models.GetNotifier()

	if msg.UserId != uint64(lead.Customer.ID) {
		log.Error(n.NotifyCustomerAboutUnreadMessage(&lead.Customer, lead, msg))
	}

	for _, seller := range lead.Shop.Sellers {
		if msg.UserId != uint64(seller.ID) {
			log.Error(n.NotifySellerAboutUnreadMessage(seller, lead, msg))
		}
	}

	if lead.Shop.NotifySupplier {
		supplier, err := models.GetUserByID(lead.Shop.SupplierID)
		if err != nil {
			log.Error(err)
			return
		}
		if msg.UserId != uint64(supplier.ID) {
			log.Error(n.NotifySellerAboutUnreadMessage(supplier, lead, msg))
		}
	}
}

func handleUserLogin(userID uint) {
	err := db.New().Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("confirmed", true).Error
	if err != nil {
		log.Error(fmt.Errorf("failed to confirm user: %v", err))
	}
}

func handleNewSession(userID uint) {
	now := time.Now()
	err := db.New().Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("last_login", now).Error
	if err != nil {
		log.Error(fmt.Errorf("failed to update last session for user %v: %v", userID, err))
	}
	err = db.New().Model(&models.Shop{}).
		Where("supplier_id = ?", userID).
		UpdateColumns(map[string]interface{}{
			"supplier_last_login": now,
			"notify_supplier":     true,
		}).Error
	if err != nil {
		log.Error(fmt.Errorf("failed to update last session in related shops for user %v: %v", userID, err))
	}
}

// notifies API about new lead
func notifyAPI(lead *models.Lead) {

	log.Debug("Notifying API about new lead (%v)", lead.ID)

	users, err := models.GetUsersForLead(lead)
	if err != nil {
		log.Error(err)
	}

	err = nats.Publish("core.lead.created", &core.NewLeadMessage{
		LeadId: uint64(lead.ID),
		Users:  users,
	})
	if err != nil {
		log.Error(err)
	}

}
