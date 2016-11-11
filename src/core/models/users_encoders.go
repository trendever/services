package models

import (
	"github.com/jinzhu/gorm"
	"proto/core"
)

// PrivateEncode encode core user model to protoUser
// if public flag is set confidential fields are filtered
func (u *User) PrivateEncode() *core.User {
	return &core.User{
		Id:        int64(u.ID),
		AvatarUrl: u.AvatarURL,

		Name:    u.Name,
		Email:   u.Email,
		Phone:   u.Phone,
		Website: u.Website,
		OptOut:  u.OptOut,
		Caption: u.Caption,
		Slogan:  u.Slogan,

		InstagramId:        u.InstagramID,
		InstagramUsername:  u.InstagramUsername,
		InstagramFullname:  u.InstagramFullname,
		InstagramAvatarUrl: u.InstagramAvatarURL,
		InstagramCaption:   u.InstagramCaption,
		HasEmail:           u.Email != "",
		HasPhone:           u.Phone != "",
		Seller:             u.IsSeller,
		Confirmed:          u.Confirmed,
	}
}

//PublicEncode converts User to the public representation
func (u *User) PublicEncode() *core.User {
	return &core.User{
		Id:        int64(u.ID),
		AvatarUrl: u.AvatarURL,
		Website:   u.Website,
		Caption:   u.Caption,
		Slogan:    u.Slogan,

		InstagramId:       u.InstagramID,
		InstagramUsername: u.InstagramUsername,
		InstagramFullname: u.InstagramFullname,
		InstagramCaption:  u.InstagramCaption,
		HasEmail:          u.Email != "",
		HasPhone:          u.Phone != "",
		Confirmed:         u.Confirmed,
		IsFake:            u.IsFake,
	}
}

//PublicEncode converts collection of users to public representation
func (u Users) PublicEncode() []*core.User {
	out := make([]*core.User, 0, len(u))
	for _, user := range u {
		out = append(out, user.PublicEncode())
	}
	return out
}

//Decode converts core.User to User
func (u User) Decode(cu *core.User) User {
	if cu == nil {
		return u
	}

	return User{
		Model: gorm.Model{
			ID: uint(cu.Id),
		},

		Name:      cu.Name,
		Email:     cu.Email,
		Phone:     cu.Phone,
		Website:   cu.Website,
		Caption:   cu.Caption,
		AvatarURL: cu.AvatarUrl,

		InstagramID:        cu.InstagramId,
		InstagramUsername:  cu.InstagramUsername,
		InstagramFullname:  cu.InstagramFullname,
		InstagramAvatarURL: cu.InstagramAvatarUrl,
		InstagramCaption:   cu.InstagramCaption,

		OptOut:      cu.OptOut,
		SuperSeller: cu.SuperSeller,
		IsSeller:    cu.Seller,
		IsFake:      cu.IsFake,
	}
}
