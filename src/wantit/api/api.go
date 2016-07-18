package api

import (
	"proto/bot"
	"proto/core"
	"utils/rpc"
	. "wantit/conf"
)

var (
	FetcherClient bot.FetcherServiceClient
	UserClient    core.UserServiceClient
	ProductClient core.ProductServiceClient
	LeadClient    core.LeadServiceClient
)

func Start() {
	fetcherConn := rpc.Connect(GetSettings().FetcherServer)
	coreConn := rpc.Connect(GetSettings().CoreServer)

	FetcherClient = bot.NewFetcherServiceClient(fetcherConn)
	UserClient = core.NewUserServiceClient(coreConn)
	ProductClient = core.NewProductServiceClient(coreConn)
	LeadClient = core.NewLeadServiceClient(coreConn)
}
