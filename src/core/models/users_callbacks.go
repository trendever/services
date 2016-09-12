package models

import (
	"core/api"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/validations"
	"github.com/ttacon/libphonenumber"
	"instagram"
	"strings"
	"utils/db"
	"utils/log"
)

//NotifyUserCreated is a notification function
var NotifyUserCreated func(u *User)

func (u *User) BeforeSave(db *gorm.DB) {
	u.validatePhone(db)
	u.fetchPreviousPhone(db)
	u.InstagramUsername = strings.ToLower(u.InstagramUsername)
}

func (u *User) AfterSave() {
	if u.previousPhone == "" && u.Phone != "" {
		go notifyUserAboutLeads(u)
		go NotifyUserCreated(u)
	}
	// we can't do it in before* callbacks because it may be long operation
	go u.CheckInstagram()
}

func (u *User) AfterUpdate() {
	go api.Publish("core.user.flush", u.ID)
}

func (u *User) AfterDelete() {
	go api.Publish("core.user.flush", u.ID)
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
func (u *User) CheckInstagram() {
	if u.ID == 0 {
		return
	}
	if u.InstagramUsername == "" {
		if u.InstagramID != 0 {
			err := db.New().Model(u).UpdateColumn("instagram_id", 0).Error
			if err != nil {
				log.Error(fmt.Errorf("failed to update instagram id for user %v: %v", u.ID, err))
			} else {
				api.Publish("core.user.flush", u.ID)
			}
		}
		return
	}

	candidates, err := api.Instagram.GetFree().SearchUsers(u.InstagramUsername)
	if err != nil {
		log.Error(fmt.Errorf("failed to search user in instagram: %v", err))
		return
	}
	var instagramInfo *instagram.SearchUserInfo
	for i := range candidates.Users {
		if candidates.Users[i].Username == u.InstagramUsername {
			instagramInfo = &candidates.Users[i]
			break
		}
	}
	updateMap := map[string]interface{}{}
	// user not found
	if instagramInfo == nil {
		if u.Name == "" {
			updateMap["name"] = u.InstagramUsername
		}
		updateMap["instagram_username"] = ""
		updateMap["instagram_id"] = 0
	} else {
		if uint64(instagramInfo.Pk) != u.InstagramID {
			updateMap["instagram_id"] = instagramInfo.Pk
		}
		if instagramInfo.ProfilePicURL != u.InstagramAvatarURL {
			avatarURL, _, err := api.ImageUploader.UploadImageByURL(instagramInfo.ProfilePicURL)
			if err == nil {
				updateMap["instagram_avatar_url"] = instagramInfo.ProfilePicURL
				updateMap["avatar_url"] = avatarURL
			} else {
				log.Error(fmt.Errorf("failed to upload new avatar for user %v: %v", u.ID, err))
			}
		}
	}
	if len(updateMap) != 0 {
		err := db.New().Model(u).UpdateColumns(updateMap).Error
		if err != nil {
			log.Error(fmt.Errorf("failed to update user %v: %v", u.ID, err))
		} else {
			api.Publish("core.user.flush", u.ID)
		}
	}
}
