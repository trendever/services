package qor

import (
	"fmt"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"utils/log"
	"core/models"
	"core/utils"
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

	log.Debug("Got token: %v", cookie.Value)

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
		log.Error(fmt.Errorf("User %v has no admin rights", user))
		return nil
	}

	return user
}
