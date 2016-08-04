package api

import (
	"api/conf"
	"google.golang.org/grpc"
	"utils/rpc"
)

// Api connections
var (
	CoreConn     *grpc.ClientConn
	AuthConn     *grpc.ClientConn
	ChatConn     *grpc.ClientConn
	PaymentsConn *grpc.ClientConn
)

// Start initializes API connections
func init() {
	settings := conf.GetSettings()

	CoreConn = rpc.Connect(settings.CoreAddr)
	AuthConn = rpc.Connect(settings.AuthAddr)
	ChatConn = rpc.Connect(settings.ChatAddr)
	PaymentsConn = rpc.Connect(settings.PaymentsAddr)
}
