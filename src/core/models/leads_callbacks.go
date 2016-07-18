package models

import (
	"utils/log"
	"core/api"
	//"core/db"
	"proto/core"
	"core/db"
	"core/notifier"
)

//BeforeSave is a gorm callback
func (l *Lead) BeforeSave() {
	if l.State == "" {
		l.State = core.LeadStatus_EMPTY.String()
	}
}

//AfterSave is gorm callback
func (l *Lead) AfterSave() {
	if !l.IsNotified && (l.State != leadStateCancelled || l.State != leadStateCompleted) {
		go func() {
			err := notifyCustomerAboutLead(l)
			if err != nil {
				log.Error(err)
			}

		}()
	}
}

func notifyCustomerAboutLead(l *Lead) (err error) {
	if l.Customer.ID == 0 {
		customer, err := GetUserByID(l.CustomerID)
		if err != nil {
			return err
		}
		l.Customer = *customer
	}

	if l.Shop.ID == 0 {
		shop, err := GetShopByID(l.ShopID)
		if err != nil {
			return err
		}
		l.Shop = *shop
	}

	if l.Customer.Phone == "" {
		return
	}
	url, err := api.GetChatURLWithToken(l.ID, l.Customer.ID)
	if err != nil {
		return
	}
	short, err := api.GetShortURL(url)
	if err != nil {
		return
	}

	err = notifier.NotifyCustomerAboutLead(l.Customer, short.URL, l, GetNotifier().NotifyBySms)
	if err != nil {
		return
	}
	return db.New().Model(l).UpdateColumn("is_notified", true).Error
}
