package views

import (
	"core/api"
	"core/models"
	"fmt"
	"proto/bot"
	"proto/chat"
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

	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "chat.message.new",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: newMessage,
	}, &nats.StanSubscription{
		Subject:        "core.notify.message",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: notifyAboutUnreadMessage,
	}, &nats.StanSubscription{
		Subject:        "auth.login",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: handleUserLogin,
	}, &nats.StanSubscription{
		Subject:        "api.new_session",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: handleNewSession,
	}, &nats.StanSubscription{
		Subject:        "coins.balance_notify",
		Group:          "core",
		DurableName:    "core",
		AckTimeout:     time.Second * 20,
		DecodedHandler: handleBalanceNotify,
	}, &nats.StanSubscription{
		Subject:        "chat.sync_event",
		Group:          "core",
		DurableName:    "core",
		AckTimeout:     time.Second * 10,
		DecodedHandler: handleSyncEvent,
	})
}

func handleSyncEvent(in *chat.Chat) bool {
	if in.SyncStatus != chat.SyncStatus_SYNCED {
		return true
	}
	lead, err := models.GetLead(0, in.Id, "Shop", "Shop.Supplier", "Customer")
	if err != nil {
		log.Error(err)
		return true
	}
	if !lead.IsNew() {
		return true
	}
	err = lead.TriggerEvent("PROGRESS", "", 0, nil)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
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

func newMessage(req *chat.NewMessageRequest) bool {
	log.Error(models.TouchLead(req.Chat.Id))

	lead, err := models.GetLead(0, req.Chat.Id, "Shop", "Shop.Supplier", "Shop.Sellers", "Customer")
	if err != nil {
		log.Error(err)
		return true
	}

	var (
		progress = false
		submit   = false
	)

	users := map[*models.User]bool{}
	for _, msg := range req.Messages {
		if msg.User.UserId != uint64(lead.Customer.ID) {
			users[&lead.Customer] = true
		} else {
			models.SendAutoAnswers(msg, lead)
		}

		isMsgAuto, err := models.IsMessageAuto(msg)
		if err != nil {
			log.Errorf("Could not check if msg is auto: %v", err)
		}

		if isMsgAuto {
			// check for progressing
			if lead.IsNew() && msg.User.Role == chat.MemberRole_CUSTOMER {
				progress = true
			} else if lead.State == "IN_PROGRESS" && msg.User.Role != chat.MemberRole_CUSTOMER {
				submit = true
			}
		}

		if lead.IsNew() {
			continue // no need in notifying people until lead is visible
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

	switch {
	case progress:
		go log.Error(lead.TriggerEvent("PROGRESS", "", 0, nil))
	case submit:
		go log.Error(submitLead(lead))
	}

	n := models.GetNotifier()
	for user := range users {
		n.NotifyUserAboutNewMessages(user, lead, req.Messages)
	}
	return true
}

func submitLead(lead *models.Lead) error {
	err := lead.TriggerEvent("SUBMIT", "", 0, nil)
	if err != nil {
		return err
	}

	if lead.Source != "comment" {
		return nil
	}

	tmpl, err := models.GetOther(models.InstagramSubmitReplyTemplate)
	if err != nil {
		return err
	}

	res, err := tmpl.Execute(lead)
	if err != nil {
		return err
	}

	renderedString, ok := res.(string)
	if !ok || renderedString <= "" {
		return fmt.Errorf("String rendered to weird shit; skipping")
	}

	log.Debug("ALL OK! Notifyin: %v", renderedString)
	var req = bot.SendDirectRequest{
		SenderId: lead.Shop.Supplier.InstagramID,
		ThreadId: lead.InstagramPk,
		Type:     bot.MessageType_ReplyComment,
		ReplyKey: "twat",
		Data:     renderedString,
	}
	err = nats.StanPublish("direct.send", &req)
	if err != nil {
		return fmt.Errorf("failed to send send comment request via nats: %v", err)
	}

	return nil
}

func notifyAboutUnreadMessage(msg *chat.Message) bool {
	lead, err := models.GetLead(0, msg.ConversationId, "Shop", "Customer")
	if err != nil {
		log.Error(err)
		return true
	}

	var count uint64
	err = db.New().Model(&models.PushToken{}).Where("user_id = ?", lead.CustomerID).Count(&count).Error
	if err != nil {
		log.Errorf("failed to determinate whether user have active push tokens: %v", err)
		return true
	}

	if count != 0 {
		return true
	}

	n := models.GetNotifier()
	log.Error(n.NotifyAboutUnreadMessage(&lead.Customer, lead, msg))
	return true
}

func handleUserLogin(userID uint) bool {
	err := db.New().Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("confirmed", true).Error
	if err != nil {
		log.Errorf("failed to confirm user: %v", err)
	}
	return true
}

func handleNewSession(userID uint) bool {
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

	return true
}
