package api

import (
	"proto/bot"
	"proto/core"
	"savetrend/conf"
	"utils/rpc"
)

// api clients
var (
	FetcherClient bot.FetcherServiceClient
	UserClient    core.UserServiceClient
	ProductClient core.ProductServiceClient
	ShopClient    core.ShopServiceClient
)

// Start api
func Start() {

	settings := conf.GetSettings()

	fetcherConn := rpc.Connect(settings.FetcherServer)
	coreConn := rpc.Connect(settings.CoreServer)

	FetcherClient = bot.NewFetcherServiceClient(fetcherConn)
	UserClient = core.NewUserServiceClient(coreConn)
	ProductClient = core.NewProductServiceClient(coreConn)
	ShopClient = core.NewShopServiceClient(coreConn)
}
