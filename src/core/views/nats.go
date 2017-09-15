package views

import (
	"core/api"
	"core/conf"
	"core/models"
	"fmt"
	"proto/accountstore"
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
	failedAutorefullTopic      = "notify_about_failed_autorefill"
	botAccountInvalidatedTopic = "bot_account_invalidated"
)

func init() {
	models.RegisterNotifyTemplates(
		failedAutorefullTopic,
		botAccountInvalidatedTopic,
	)

	nats.StanSubscribe(&nats.StanSubscription{
		Subject:        "chat.message.new",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: newMessage,
	}, &nats.StanSubscription{
		Subject:        "chat.unanswered",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: handleUnansweredMessages,
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
	}, &nats.StanSubscription{
		Subject:        "accountstore.notify",
		Group:          "core",
		DurableName:    "core",
		DecodedHandler: handleAccountstoreNotify,
	})
}

func handleAccountstoreNotify(acc *accountstore.Account) bool {
	if acc.Valid || acc.OwnerId == 0 {
		return true
	}
	log.Error(models.GetNotifier().NotifyUserByID(acc.OwnerId, botAccountInvalidatedTopic, map[string]interface{}{
		"account": acc,
		"url": api.GetShortURL(
			api.AddUserToken(conf.GetSettings().URL.ConnectBot, acc.OwnerId),
		),
	}))
	return true
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
		// @TODO technically it's possible to get second event if sync will be reenabled after error...
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

	//Возможные переходы
	var (
		progress = false
		submit   = false
	)

	//Создаем мапу для уведомлений
	users := map[*models.User]bool{}

	for _, msg := range req.Messages {
		//Если юзер не клиент
		if msg.User.UserId != uint64(lead.Customer.ID) {
			//пихаем клиента в массив уведомлений
			users[&lead.Customer] = true
		} else {
			//Если сообщение от кастомера то шлем автоответы
			models.SendAutoAnswers(msg, lead)
		}

		isMsgAuto, err := models.IsMessageAuto(msg)
		if err != nil {
			log.Errorf("Could not check if msg is auto: %v", err)
		}

		if lead.IsNew() && msg.User.Role == chat.MemberRole_CUSTOMER {
			progress = true
		}

		if !isMsgAuto {
			// check for progressing
			if lead.State == "IN_PROGRESS" &&
				msg.User.Role != chat.MemberRole_CUSTOMER && msg.User.Role != chat.MemberRole_SYSTEM {
				submit = true
			}
		}

		if lead.IsNew() {
			continue // no need in notifying people until lead is visible
		}

		//Если юзер не поставщик
		if msg.User.UserId != uint64(lead.Shop.SupplierID) {
			//Уведомляем поставщика
			users[&lead.Shop.Supplier] = true
		}

		//Также уведомляем всех селлеров
		for _, seller := range lead.Shop.Sellers {
			// Кроме того, который это сообщение отправил
			if msg.User.UserId != uint64(seller.ID) {
				users[seller] = true
			}
		}
	}

	//Меняем стейт
	switch {
	case progress:
		go log.Error(lead.TriggerEvent("PROGRESS", "", 0, nil))
	case submit:
		if lead.Source == "comment" {
			err := models.SubmitCommentReply(lead)
			if err != nil {
				log.Errorf("failed to send submit reply comment: %v", err)
				return false
			}
		}
		go log.Error(lead.TriggerEvent("SUBMIT", "", 0, nil))
	}

	//Уведомляем чувачков из мапы
	n := models.GetNotifier()
	for user := range users {
		n.NotifyUserAboutNewMessages(user, lead, req.Messages)
	}
	return true
}

func handleUnansweredMessages(notify *chat.UnansweredNotify) bool {
	preload := []string{"Shop", "Customer"}
	if !notify.ForUser {
		preload = []string{"Customer", "Shop", "Shop.Supplier", "Shop.Sellers"}
	}
	lead, err := models.GetLead(0, notify.ChatId, preload...)
	if err != nil {
		log.Error(err)
		return true
	}

	n := models.GetNotifier()
	if notify.ForUser {
		var count uint64
		err = db.New().Model(&models.PushToken{}).Where("user_id = ?", lead.CustomerID).Count(&count).Error
		if err != nil {
			log.Errorf("failed to determinate whether user have active push tokens: %v", err)
			return true
		}

		if count != 0 {
			return true
		}

		log.Error(n.NotifyAboutUnansweredMessages(&lead.Customer, lead, notify.Count, notify.Group, notify.Messages))
	} else {
		log.Error(n.NotifyAboutUnansweredMessages(&lead.Shop.Supplier, lead, notify.Count, notify.Group, notify.Messages))
		for _, seller := range lead.Shop.Sellers {
			log.Error(n.NotifyAboutUnansweredMessages(seller, lead, notify.Count, notify.Group, notify.Messages))
		}
	}
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
	tx := db.NewTransaction()
	err := tx.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("last_login", now).Error
	if err != nil {
		log.Errorf("failed to update last session for user %v: %v", userID, err)
		tx.Rollback()
		return false
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
	// reset init plan expiration time on first login of supplier
	err = tx.Model(&models.Shop{}).
		Where("supplier_id = ?", userID).
		Where("plan_id = ?", plan.ID).
		Where("supplier_last_login = ? OR supplier_last_login IS NULL", time.Time{}).
		UpdateColumns(updateMap).Error
	if err != nil {
		tx.Rollback()
		log.Errorf("failed to update init plan expiration time for user %v: %v", userID, err)
		return false
	}

	err = tx.Model(&models.Shop{}).
		Where("supplier_id = ?", userID).
		UpdateColumns(map[string]interface{}{
			"supplier_last_login": now,
			"notify_supplier":     true,
		}).Error
	if err != nil {
		tx.Rollback()
		log.Errorf("failed to update last session in related shops for user %v: %v", userID, err)
		return false
	}
	err = tx.Commit().Error
	if err != nil {
		log.Errorf("db transaction failed when handling new session of user %v: %v", userID, err)
		return false
	}

	return true
}
