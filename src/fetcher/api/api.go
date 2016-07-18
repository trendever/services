package api

import (
	"utils/rpc"
	"google.golang.org/grpc"
	. "fetcher/conf"
)

var GrpcServer *grpc.Server

func Start() {
	GrpcServer = rpc.Serve(GetSettings().RPC)
}
