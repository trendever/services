package api

import (
	"utils/rpc"
	"google.golang.org/grpc"
	"api/conf"
)

// Api connections
var (
	CoreConn *grpc.ClientConn
	AuthConn *grpc.ClientConn
	ChatConn *grpc.ClientConn
)

// Start initializes API connections
func init() {
	settings := conf.GetSettings()

	CoreConn = rpc.Connect(settings.CoreAddr)
	AuthConn = rpc.Connect(settings.AuthAddr)
	ChatConn = rpc.Connect(settings.ChatAddr)
}
