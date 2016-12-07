package views

import (
	"core/api"
	"core/models"
	"fmt"
	"proto/chat"
	"proto/core"
	"proto/payment"
	"proto/trendcoin"
	"time"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"
)

const (
	failedAutorefullTopic = "notify_about_failed_autorefill"
)

func init() {
	models.RegisterNotifyTemplate(failedAutorefullTopic)

	nats.Subscribe(&nats.Subscription{
		Subject: "chat.message.new",
		Group:   "core",
		Handler: newMessage,
	})
	nats.Subscribe(&nats.Subscription{
		Subject: "core.notify.message",
		Group:   "core",
		Handler: notifyAboutUnreadMessage,
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
	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "coins.balance_notify",
		Group:          "core",
		DurableName:    "core",
		AckTimeout:     time.Second * 20,
		DecodedHandler: handleBalanceNotify,
	})
}

func handleBalanceNotify(notify *trendcoin.BalanceNotify) bool {
	log.Debug("new balance notify: %+v", notify)
	// autorefill failed, disable it
	if notify.Failed {
		err := db.New().Delete(&models.AutorefillInfo{UserID: notify.UserId}).Error
		if err != nil {
			log.Errorf("failed to delete autorefill info after payment fail: %v", err)
			return false
		}
		log.Error(models.GetNotifier().NotifyUserByID(notify.UserId, failedAutorefullTopic, map[string]interface{}{}))
		return true
	}

	var autorefill models.AutorefillInfo
	res := db.New().First(&autorefill, "user_id = ?", notify.UserId)
	if res.RecordNotFound() {
		log.Debug("user %v has no autorefill", notify.UserId)
		return true
	}
	if res.Error != nil {
		log.Errorf("failed to load autorefill info: %v", res.Error)
		return false
	}

	// that was autorefill event
	if notify.Autorefill {
		err := db.New().Model(&autorefill).UpdateColumn("in_progress", false).Error
		if err != nil {
			log.Errorf("failed to update autorefill status: %v", err)
			return false
		}
		autorefill.InProgress = false
	}
	// positive balance or autorefill is still in progress
	if notify.Balance >= 0 || autorefill.InProgress {
		return true
	}

	// autorefill time
	var offer models.CoinsOffer
	res = db.New().First(&offer, "id = ?", autorefill.CoinsOffer)

	if res.RecordNotFound() {
		err := db.New().Delete(&autorefill).Error
		if err != nil {
			log.Errorf("failed to delete invalid autorefill info")
			return false
		}
		log.Error(models.GetNotifier().NotifyUserByID(notify.UserId, failedAutorefullTopic, map[string]interface{}{}))
		return true
	}
	if res.Error != nil {
		log.Errorf("failed to load motetization plan: %v", res.Error)
		return false
	}

	currency, ok := payment.Currency_value[offer.Currency]
	if !ok {
		// wtf? outdated offer or inconsistent db?
		log.Errorf("coins offer %v uses unknown currency %v", offer.ID, offer.Currency)
		log.Error(models.GetNotifier().NotifyUserByID(notify.UserId, failedAutorefullTopic, map[string]interface{}{}))
		return true
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	payResp, err := api.PaymentsServiceClient.BuyAsync(ctx, &payment.BuyAsyncRequest{
		Data: &payment.OrderData{
			Amount:      uint64(offer.Amount),
			Currency:    payment.Currency(currency),
			Gateway:     "payture_ewallet",
			ServiceName: "coins_refill",
			ServiceData: fmt.Sprintf(`{"user_id": %v, "amount": %v, "autorefill": true}`, notify.UserId, offer.Amount),
			Comment:     fmt.Sprintf("%v trendcoins autorefill", offer.Amount),
		},
		User: &payment.UserInfo{
			UserId: notify.UserId,
		},
	})
	if err != nil {
		return false
	}

	if payResp.Error != payment.Errors_OK {
		log.Errorf("payments service returned error: %v", payResp.ErrorMessage)
		return false
	}

	err = db.New().Model(&autorefill).UpdateColumn("in_progress", true).Error
	if err != nil {
		log.Errorf("failed to update autorefill status: %v", err)
		return false
	}

	return true
}

func newMessage(req *chat.NewMessageRequest) {
	log.Error(models.TouchLead(req.Chat.Id))

	lead, err := models.GetLead(0, req.Chat.Id, "Shop", "Shop.Supplier", "Shop.Sellers", "Customer")
	if err != nil {
		log.Error(err)
		return
	}
	newLead := lead.IsNew()

	users := map[*models.User]bool{}
	for _, msg := range req.Messages {
		if msg.User.UserId != uint64(lead.Customer.ID) {
			users[&lead.Customer] = true
		} else {
			models.SendAutoAnswers(msg, lead)
		}

		if newLead {
			continue
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

func notifyAboutUnreadMessage(msg *chat.Message) {
	lead, err := models.GetLead(0, msg.ConversationId, "Shop", "Customer")
	if err != nil {
		log.Error(err)
		return
	}

	var count uint64
	err = db.New().Model(&models.PushToken{}).Where("user_id = ?", lead.CustomerID).Count(&count).Error
	if err != nil {
		log.Errorf("failed to determinate whether user have active push tokens: %v", err)
		return
	}

	n := models.GetNotifier()
	log.Error(n.NotifyAboutUnreadMessage(&lead.Customer, lead, msg))
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

// NotifyAboutLeadEvent notifies about lead event via NATS, changes related conversation status
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
