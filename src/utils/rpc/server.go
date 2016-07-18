package rpc

import (
	"utils/log"
	"google.golang.org/grpc"
	"net"
)

//Serve starts grpc.Server and returns it instance
func Serve(addr string) *grpc.Server {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	go server.Serve(lis)

	return server
}
