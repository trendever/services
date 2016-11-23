package views

import (
	"api/api"
	"api/soso"
	"proto/accountstore"
)

var accountStoreServiceClient = accountstore.NewAccountStoreServiceClient(api.AccountStoreConn)

func init() {
	SocketRoutes = append(SocketRoutes,
		soso.Route{"account", "retrieve", RetrieveAccounts},
	)
}

func RetrieveAccounts(c *soso.Context) {

}
