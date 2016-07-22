package rpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
	"utils/log"
)

//Connect makes new connection to a grpc server
func Connect(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

//DefaultContext returns context with default timeout
func DefaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}
