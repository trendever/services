package models

import (
	"core/api"
	"core/conf"
	"fmt"
	"github.com/jinzhu/gorm"
	"utils/db"
	"utils/log"
)

// User model
type User struct {
	gorm.Model

	Name      string
	Email     string
	Phone     string `gorm:"index"`
	Website   string
	AvatarURL string `gorm:"text"`

	// Instagram fields
	InstagramID        uint64
	InstagramUsername  string `gorm:"index"`
	InstagramFullname  string
	InstagramAvatarURL string

	// instagram calls this `biography`. can be really long
	InstagramCaption string `gorm:"type:text"`

	// Just like InstagramCaption, but rw for qor
	Caption string `gorm:"type:text"`
	// Short status-like string
	Slogan string

	OptOut bool

	// access to qor
	IsAdmin bool `sql:"default:false"`
	// IsScout indicate this user is our user for collecting trends
	IsScout bool `sql:"default:false"`
	// ability to be responsible for chats
	IsSeller bool `sql:"default:false"`
	// ability to join any chat
	SuperSeller bool `sql:"default:false"`

	previousPhone string
	// true if the user was logged in at least once
	Confirmed bool `sql:"default:false"`
}

// SystemUser is used if we need to send a message from system
var SystemUser User

// LoadOrCreateSystemUser func
func LoadOrCreateSystemUser() error {
	name := conf.GetSettings().SystemUser
	res := db.New().Find(&SystemUser, "name = ?", name)
	if res.RecordNotFound() {
		log.Warn("System user with name %v not found, creating new one", name)
		SystemUser.Name = name
		return db.New().Create(&SystemUser).Error
	}
	return res.Error
}

//Users is an array of users
type Users []User

// returns something that can be used like first name
func (u User) getName() string {
	switch {
	case u.Name != "":
		return u.Name
	case u.InstagramFullname != "" && u.InstagramUsername != u.InstagramFullname:
		return fmt.Sprintf("%v (@%v)", u.InstagramFullname, u.InstagramUsername)
	case u.InstagramUsername != "":
		return fmt.Sprintf("@%s", u.InstagramUsername)
	default:
		return ""
	}
}

// returns something can be used like an unique token
func (u User) getAddr() string {
	switch {
	case u.Phone != "":
		return u.Phone
	case u.Email != "":
		return u.Email
	default:
		return "unknown"
	}

}

// Stringify generates pretty-name generally for qor
// qor really wants this to be non-empty, that's what all the buzz about
func (u User) Stringify() string {
	name := u.getName()
	addr := u.getAddr()

	switch {
	case u.ID == 0:
		return "Deleted user"
	case name != "":
		return name
	case addr != "":
		return addr
	default:
		return fmt.Sprintf("User id=#%v", u.ID)
	}
}

// DisplayName returns name should be displayed in qor interface
func (u User) DisplayName() string {
	return u.getName()
}

// ResourceName returns qor resource name
func (u User) ResourceName() string {
	return "Users"
}

// TableName for this model
func (u User) TableName() string {
	return "users_user"
}

// UserHasEmail is a helper function: check if user has an email
// in future, it should also check if this email is confirmed
func UserHasEmail(userID uint) (bool, error) {
	user := User{}

	err := db.
		New().
		Where("id = ?", userID).
		Find(&user).
		Error

	if err != nil {
		return false, err
	}

	return user.Email != "", nil
}

//GetName returns printable name for qor usage
func (u *User) GetName() string {
	switch {
	case u.InstagramUsername != "":
		return u.InstagramUsername
	case u.Name != "":
		return u.Name
	}
	return "User"
}

//AfterUpdate is a gorm callback
func (u *User) AfterUpdate() {
	go api.Publish("core.user.flush", u.ID)
}
