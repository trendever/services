package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/accountstore"
	"utils/rpc"
)

var accountStoreServiceClient = accountstore.NewAccountStoreServiceClient(api.AccountStoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "account", RetrieveAccount},
		soso.Route{"add", "account", AddAccount},

		//		soso.Route{"account", "list", RetrieveAccounts},
		//		soso.Route{"account", "add", AddAccount},
		//		soso.Route{"account", "invalidate", MarkInvalid},
		//		soso.Route{"account", "confirm", Confirm},
	)
}

// RetrieveAccount gets this account bot
func RetrieveAccount(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	// get current user instagram ID
	user, err := GetUser(c.Token.UID, false)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	instagramID := user.InstagramId

	if instagramID == 0 {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Zero instagram ID"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := accountStoreServiceClient.Get(ctx, &accountstore.GetRequest{
		InstagramId: instagramID,
	})
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"account": resp.Account,
		"found":   resp.Found,
	})

}

func AddAccount(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	// get current user instagram ID
	user, err := GetUser(c.Token.UID, false)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	instagramUsername := user.InstagramUsername

	if instagramUsername == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Zero instagram username"))
		return
	}

	password, ok := c.RequestMap["password"].(string)
	if !ok || password == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("No password supplied"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := accountStoreServiceClient.Add(ctx, &accountstore.AddRequest{
		InstagramUsername: instagramUsername,
		Password:          password,
	})
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success":   true,
		"need_code": resp.NeedCode,
	})

}

func Confirm(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	// get current user instagram ID
	user, err := GetUser(c.Token.UID, false)
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	instagramUsername := user.InstagramUsername

	if instagramUsername == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Zero instagram username"))
		return
	}

	code, ok := c.RequestMap["code"].(string)
	if !ok || code == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("No code supplied"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	_, err = accountStoreServiceClient.Confirm(ctx, &accountstore.ConfirmRequest{
		InstagramUsername: instagramUsername,
		Code:              code,
	})
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": true,
	})

}
