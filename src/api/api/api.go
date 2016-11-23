package api

import (
	"api/conf"
	"google.golang.org/grpc"
	"utils/rpc"
)

// Api connections
var (
	CoreConn         *grpc.ClientConn
	AuthConn         *grpc.ClientConn
	ChatConn         *grpc.ClientConn
	SMSConn          *grpc.ClientConn
	PaymentsConn     *grpc.ClientConn
	CoinsConn        *grpc.ClientConn
	CheckerConn      *grpc.ClientConn
	AccountStoreConn *grpc.ClientConn
)

// Start initializes API connections
func init() {
	settings := conf.GetSettings()

	CoreConn = rpc.Connect(settings.API.Core)
	AuthConn = rpc.Connect(settings.API.Auth)
	ChatConn = rpc.Connect(settings.API.Chat)
	SMSConn = rpc.Connect(settings.API.SMS)
	PaymentsConn = rpc.Connect(settings.API.Payments)
	CoinsConn = rpc.Connect(settings.API.Coins)
	CheckerConn = rpc.Connect(settings.API.Checker)
	AccountStoreConn = rpc.Connect(settings.API.AccountStore)
}
