package exteral

import (
	"proto/core"
	"push/config"
	"utils/rpc"
)

var PushTokensServiceClient core.PushTokensServiceClient

func Init() {
	pushConn := rpc.Connect(config.Get().PushTokensServer)
	PushTokensServiceClient = core.NewPushTokensServiceClient(pushConn)
}
