package views

import (
	"api/api"
	"api/soso"
	"errors"
	"net/http"
	"proto/accountstore"
	"proto/core"
	"utils/rpc"
)

var accountStoreServiceClient = accountstore.NewAccountStoreServiceClient(api.AccountStoreConn)

func init() {
	SocketRoutes = append(
		SocketRoutes,
		soso.Route{"retrieve", "account", RetrieveAccount},
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

	var instagramID uint64

	{ // get current user instagram ID
		ctx, cancel := rpc.DefaultContext()
		defer cancel()
		resp, err := userServiceClient.ReadUser(ctx, &core.ReadUserRequest{
			Id: c.Token.UID,
		})
		if err != nil {
			c.ErrorResponse(http.StatusBadRequest, soso.LevelError, err)
			return
		}

		instagramID = resp.User.InstagramId
	}

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
