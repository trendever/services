package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/accountstore"
	"utils/log"
	"utils/rpc"

	"golang.org/x/net/context"
)

var accountStoreServiceClient = accountstore.NewAccountStoreServiceClient(api.AccountStoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"add", "account", AddBot},
		soso.Route{"list", "account", ListAccounts},
		soso.Route{"confirm", "account", Confirm},
		//		soso.Route{"account", "invalidate", MarkInvalid},
	)
}

// Confirm user -- apply instagram checkpoint
func Confirm(c *soso.Context) {
	if c.Token == nil {
		c.ErrorResponse(403, soso.LevelError, errors.New("User not authorized"))
		return
	}

	instagramUsername, _ := c.RequestMap["instagram_username"].(string)

	if instagramUsername == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("Zero instagram username"))
		return
	}

	code, ok := c.RequestMap["code"].(string)
	if !ok || code == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("No code supplied"))
		return
	}

	password, ok := c.RequestMap["password"].(string)
	if !ok || code == "" {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("No password supplied"))
		return
	}

	_, err := accountStoreServiceClient.Confirm(context.Background(), &accountstore.ConfirmRequest{
		InstagramUsername: instagramUsername,
		Code:              code,
		Password:          password,
	})
	if err != nil {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
		return
	}

	c.SuccessResponse(map[string]interface{}{
		"success": true,
	})

}

// ListAccounts returns list of available accs
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

	withInvalids, _ := c.RequestMap["with_invalids"].(bool)
	withNonOwned, _ := c.RequestMap["with_non_owned"].(bool)
	showPrivate, _ := c.RequestMap["show_private"].(bool)

	req := accountstore.SearchRequest{
		IncludeInvalids: withInvalids,
		OwnerId:         c.Token.UID,
		HidePrivate:     !user.IsAdmin || !showPrivate,
	}

	roleName, ok := c.RequestMap["role"].(string)
	if ok && user.IsAdmin && roleName > "" {
		role, ok := accountstore.Role_value[roleName]
		if !ok {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("unknown role"))
			return
		}
		req.Roles = []accountstore.Role{accountstore.Role(role)}
	}

	if user.IsAdmin && withNonOwned {
		req.OwnerId = 0
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

// AddBot routine
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

	if !user.IsAdmin && role != int32(accountstore.Role_User) {
		c.ErrorResponse(http.StatusBadRequest, soso.LevelError, errors.New("only admins can add bots"))
		return
	}

	preferEmail, _ := c.RequestMap["prefer_email"].(bool)
	proxy, _ := c.RequestMap["proxy"].(string)

	resp, err := accountStoreServiceClient.Add(context.Background(), &accountstore.AddRequest{
		InstagramUsername: username,
		Password:          password,
		Role:              accountstore.Role(role),
		OwnerId:           c.Token.UID,
		PreferEmail:       preferEmail,
		Proxy:             proxy,
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
