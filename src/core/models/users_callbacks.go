package models

import (
	"core/db"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"github.com/ttacon/libphonenumber"
	"strings"
	"utils/log"
)

//NotifyUserCreated is a notification function
var NotifyUserCreated func(u *User)

// BeforeSave saves edited phone number (and edit it if neccessary)
func (u *User) BeforeSave(db *gorm.DB) {
	u.validatePhone(db)
	u.fetchPreviousPhone(db)

	u.InstagramUsername = strings.ToLower(u.InstagramUsername)

}

func (u *User) fetchPreviousPhone(db *gorm.DB) {
	origin := &User{}
	if err := db.Model(&User{}).Select("phone").Find(origin, u.ID); err == nil {
		u.previousPhone = origin.Phone
	}
}

func (u *User) validatePhone(db *gorm.DB) {
	if u.Phone != "" {
		newPhone, err := libphonenumber.Parse(u.Phone, "")
		correct := libphonenumber.IsValidNumber(newPhone)

		switch {
		case err != nil || !correct:
			db.AddError(validations.NewError(u, "Phone", err.Error()))
		case !correct:
			db.AddError(validations.NewError(u, "Phone", "Uncorrect phone number"))
		default:
			u.Phone = libphonenumber.Format(newPhone, libphonenumber.E164)
		}
	}
}

//AfterSave is gorm callback
func (u *User) AfterSave() {
	if u.previousPhone == "" && u.Phone != "" {
		go notifyUserAboutLeads(u)
		go NotifyUserCreated(u)
	}
}

func notifyUserAboutLeads(user *User) {
	leads := []*Lead{}
	scope := db.New().
		Model(&Lead{}).
		//we want to notify the user only about leads which didn't finish and didn't notified before this
		Where("customer_id = ? AND is_notified = ? AND state not in (?)", user.ID, false, []string{leadStateCancelled, leadStateCompleted}).
		Preload("Shop").
		Preload("Shop.Sellers").
		Preload("Shop.Supplier").
		Preload("Customer").
		Find(&leads)
	if scope.Error != nil {
		log.Error(scope.Error)
		return
	}

	for _, lead := range leads {
		err := notifyCustomerAboutLead(lead)
		if err != nil {
			log.Error(err)
		}
		SendStatusMessage(lead.ConversationID, "customer.phone.added", "")
		if err != nil {
			//just log, not critical
			log.Error(err)
		}
	}
}
