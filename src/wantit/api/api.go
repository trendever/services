package api

import (
	"proto/bot"
	"proto/core"
	"utils/rpc"
	. "wantit/conf"
)

var (
	FetcherClient   bot.FetcherServiceClient
	SaveTrendClient bot.SaveTrendServiceClient
	UserClient      core.UserServiceClient
	ProductClient   core.ProductServiceClient
	LeadClient      core.LeadServiceClient
	ShopClient      core.ShopServiceClient
)

func Start() {
	fetcherConn := rpc.Connect(GetSettings().FetcherServer)
	coreConn := rpc.Connect(GetSettings().CoreServer)
	saveTrendConn := rpc.Connect(GetSettings().SaveTrendServer)

	FetcherClient = bot.NewFetcherServiceClient(fetcherConn)
	SaveTrendClient = bot.NewSaveTrendServiceClient(saveTrendConn)
	UserClient = core.NewUserServiceClient(coreConn)
	ProductClient = core.NewProductServiceClient(coreConn)
	LeadClient = core.NewLeadServiceClient(coreConn)
	ShopClient = core.NewShopServiceClient(coreConn)
}
