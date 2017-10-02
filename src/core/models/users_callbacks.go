package models

import (
	"common/db"
	"common/log"
	"core/api"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"proto/checker"
	"proto/trendcoin"
	"strings"
	"utils/nats"
	"utils/phone"
	"utils/rpc"
)

func (t Telegram) AfterSave(db *gorm.DB) {
	err := db.Exec("UPDATE users_user SET has_telegram = EXISTS (SELECT 1 FROM telegrams WHERE user_id = ? AND confirmed) WHERE id = ?", t.UserID, t.UserID).Error
	if err != nil {
		db.AddError(err)
	}
}

func (t Telegram) AfterDelete(db *gorm.DB) {
	err := db.Exec("UPDATE users_user SET has_telegram = EXISTS (SELECT 1 FROM telegrams WHERE user_id = ? AND confirmed) WHERE id = ?", t.UserID, t.UserID).Error
	if err != nil {
		db.AddError(err)
	}
}

func (u *User) BeforeSave(db *gorm.DB) {
	u.validatePhone(db)
	u.fetchPreviousPhone(db)
	u.Name = strings.Trim(u.Name, " \t\n@")
	u.InstagramUsername = strings.ToLower(strings.Trim(u.InstagramUsername, " \n\t"))
	if u.InstagramUsername == "" {
		u.InstagramID = 0
	}
}

func (u *User) AfterCommit() {
	if u.previousPhone == "" && u.Phone != "" {
		go notifyUserAboutLeads(u)
		go NotifyUserCreated(u)
	}
	if u.InstagramUsername != "" {
		go func() {
			u.InitInstagramCheck()
		}()
	}
}

func (u *User) AfterUpdate() {
	go nats.Publish("core.user.flush", u.ID)
}

func (u *User) AfterDelete() {
	go nats.Publish("core.user.flush", u.ID)
}

func (u *User) LoadExternals(db *gorm.DB) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := api.TrendcoinServiceClient.Balance(ctx, &trendcoin.BalanceRequest{
		UserId: uint64(u.ID),
	})
	if err != nil {
		db.AddError(fmt.Errorf("failed to load balance: %v", err))
		return
	}
	if res.Error != "" {
		db.AddError(fmt.Errorf("failed to load balance: %v", res.Error))
		return
	}
	u.Balance = res.Balance
}

func (u *User) fetchPreviousPhone(db *gorm.DB) {
	origin := &User{}
	if err := db.Model(&User{}).Select("phone").Find(origin, u.ID).Error; err == nil {
		u.previousPhone = origin.Phone
	}
}

func (u *User) validatePhone(db *gorm.DB) {
	if u.Phone != "" {
		phoneNumber, err := phone.CheckNumber(u.Phone, "")

		if err != nil {
			db.AddError(validations.NewError(u, "Phone", err.Error()))
		}

		u.Phone = phoneNumber
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

// checks if instagram user exists and updates instagramID and avatar
func (u *User) InitInstagramCheck() {
	if u.ID == 0 {
		return
	}
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	api.CheckerServiceClient.Check(ctx, &checker.CheckRequest{Ids: []uint64{uint64(u.ID)}})
}
