package typemap

import (
	"proto/core"
	"proto/push"
)

// @TODO use protobuf import instead of this
// actuality --go_out=import_path option of protoc do nothing, so generated code for imports is broken

var ServiceToTokenType map[push.ServiceType]core.TokenType
var TokenTypeToService map[core.TokenType]push.ServiceType

func init() {
	ServiceToTokenType = map[push.ServiceType]core.TokenType{
		push.ServiceType_APN: core.TokenType_Iphone,
		push.ServiceType_FCM: core.TokenType_Android,
	}
	for k, v := range ServiceToTokenType {
		TokenTypeToService[v] = k
	}
}
