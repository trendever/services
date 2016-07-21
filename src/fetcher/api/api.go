package api

import (
	. "fetcher/conf"
	"google.golang.org/grpc"
	"utils/rpc"
)

var GrpcServer *grpc.Server

func Start() {
	GrpcServer = rpc.Serve(GetSettings().RPC)
}
