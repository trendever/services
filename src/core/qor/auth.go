package qor

import (
	"core/models"
	"core/utils"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"utils/log"
)

const (
	tokenName = "_te_token"
	loginURL  = "/signup"
	logoutURL = "/logout"
)

// Auth brings our common auth to qor
type Auth struct{}

// LoginURL returns login URL
func (Auth) LoginURL(c *admin.Context) string {
	return loginURL
}

// LogoutURL returns logout URL
func (Auth) LogoutURL(c *admin.Context) string {
	return logoutURL
}

// GetCurrentUser returns current user
func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {

	cookie, err := c.Request.Cookie(tokenName)
	if err != nil {
		log.Error(err)
		return nil
	}

	tokenData, err := utils.GetTokenData(cookie.Value)
	if err != nil {
		log.Error(err)
		return nil
	}

	user, err := models.GetUserByID(uint(tokenData.UID))
	if err != nil {
		log.Error(err)
		return nil
	}

	if !user.IsAdmin {
		log.Errorf("User %v has no admin rights", user)
		return nil
	}

	return user
}
