package views

import (
	"core/models"
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

	lead, err := models.GetLead(0, req.Chat.Id, "Shop", "Shop.Supplier", "Shop.Sellers", "Customer")
	if err != nil {
		log.Error(err)
		return
	}

	users := map[*models.User]bool{}
	for _, msg := range req.Messages {
		if msg.User.UserId != uint64(lead.Customer.ID) {
			users[&lead.Customer] = true
		} else {
			models.SendAutoAnswers(msg, lead)
		}

		if msg.User.UserId != uint64(lead.Shop.SupplierID) {
			users[&lead.Shop.Supplier] = true
		}
		for _, seller := range lead.Shop.Sellers {
			if msg.User.UserId != uint64(seller.ID) {
				users[seller] = true
			}
		}
	}
	n := models.GetNotifier()
	for user := range users {
		n.NotifyUserAboutNewMessages(user, lead, req.Messages)
	}
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
		log.Errorf("failed to confirm user: %v", err)
	}
}

func handleNewSession(userID uint) {
	now := time.Now()
	err := db.New().Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("last_login", now).Error
	if err != nil {
		log.Errorf("failed to update last session for user %v: %v", userID, err)
	}

	plan := &models.InitialPlan
	updateMap := map[string]interface{}{
		"plan_id":          plan.ID,
		"suspended":        false,
		"auto_renewal":     false,
		"last_plan_update": time.Now(),
		"plan_expires_at":  time.Time{},
	}
	if plan.SubscriptionPeriod != 0 {
		updateMap["plan_expires_at"] = time.Now().Add(models.PlansBaseDuration * time.Duration(plan.SubscriptionPeriod))
	}
	err = db.New().Model(&models.Shop{}).
		Where("supplier_id = ?", userID).
		Where("supplier_last_login = ? OR supplier_last_login IS NULL", time.Time{}).
		UpdateColumns(updateMap).Error

	err = db.New().Model(&models.Shop{}).
		Where("supplier_id = ?", userID).
		UpdateColumns(map[string]interface{}{
			"supplier_last_login": now,
			"notify_supplier":     true,
		}).Error
	if err != nil {
		log.Errorf("failed to update last session in related shops for user %v: %v", userID, err)
	}
}

// notifies about lead event via NATS, changes related conversation status
func NotifyAboutLeadEvent(lead *models.Lead, event string) {

	log.Debug("Notifying about lead %v event", lead.ID)

	users, err := models.GetUsersForLead(lead)
	if err != nil {
		log.Errorf("failed to get related users for lead %v: %v", lead.ID, err)
	}

	err = nats.Publish("core.lead.event", &core.LeadEventMessage{
		LeadId: uint64(lead.ID),
		Users:  users,
		Event:  event,
	})
	if err != nil {
		log.Errorf("failed to publush core.lead.event: %v", err)
	}

	chatStatus := "new"
	switch lead.State {
	case core.LeadStatus_NEW.String(), core.LeadStatus_EMPTY.String():

	case core.LeadStatus_CANCELLED.String():
		chatStatus = "cancelled"
	default:
		chatStatus = "active"
	}
	err = nats.Publish("chat.conversation.set_status", &chat.SetStatusMessage{
		ConversationId: lead.ConversationID,
		Status:         chatStatus,
	})
	if err != nil {
		log.Errorf("failed to publush chat.conversation.set_status: %v", err)
	}
}
