package models

import (
	"proto/core"
	"utils/db"
	"utils/log"
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
	err = GetNotifier().NotifyCustomerAboutLead(&l.Customer, l)
	if err != nil {
		return
	}
	return db.New().Model(l).UpdateColumn("is_notified", true).Error
}
