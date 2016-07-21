package rpc

import (
	"google.golang.org/grpc"
	"net"
	"utils/log"
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
