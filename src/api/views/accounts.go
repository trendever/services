package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/accountstore"
	"utils/log"
	"utils/rpc"
)

var accountStoreServiceClient = accountstore.NewAccountStoreServiceClient(api.AccountStoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "account", RetrieveAccount},
		soso.Route{"add", "account", AddAccount},

		// admins methods
		soso.Route{"list", "account", ListAccounts},
		soso.Route{"add_bot", "account", AddBot},

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
		Role:              accountstore.Role_User,
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

func ListAccounts(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("User not authorized"))
		return
	}
	user, err := GetUser(c.Token.UID, false)
	log.Debug("user: %v", log.IndentEncode(user))
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if !user.IsAdmin {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("Only admins can do it"))
		return
	}

	withInvalids, _ := c.RequestMap["with_invalids"].(bool)

	req := accountstore.SearchRequest{
		IncludeInvalids: withInvalids,
	}
	roleName, ok := c.RequestMap["role"].(string)
	if ok {
		role, ok := accountstore.Role_value[roleName]
		if !ok {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("unknown role"))
			return
		}
		req.Roles = []accountstore.Role{accountstore.Role(role)}
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	res, err := accountStoreServiceClient.Search(ctx, &req)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	c.SuccessResponse(res.Accounts)
}

func AddBot(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("User not authorized"))
		return
	}
	user, err := GetUser(c.Token.UID, false)
	if err != nil {
		c.ErrorResponse(http.StatusInternalServerError, soso.LevelError, err)
		return
	}
	if !user.IsAdmin {
		c.ErrorResponse(http.StatusForbidden, soso.LevelError, errors.New("Only admins can do it"))
		return
	}

	username, _ := c.RequestMap["username"].(string)
	password, _ := c.RequestMap["password"].(string)
	if username == "" || password == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("empty username or password"))
		return
	}

	roleName, _ := c.RequestMap["role"].(string)
	role, ok := accountstore.Role_value[roleName]
	if !ok {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("unknown role"))
		return
	}

	if role == int32(accountstore.Role_User) {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("you can not add user like this"))
		return
	}

	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	resp, err := accountStoreServiceClient.Add(ctx, &accountstore.AddRequest{
		InstagramUsername: username,
		Password:          password,
		Role:              accountstore.Role(role),
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
