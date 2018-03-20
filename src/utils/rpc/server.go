package rpc

import (
	"common/log"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	*grpc.Server
	listener net.Listener
}

//Serve starts grpc.Server and returns it instance
func MakeServer(addr string) Server {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := Server{
		Server:   grpc.NewServer(),
		listener: lis,
	}

	return server
}

func (s Server) StartServe() {
	go s.Server.Serve(s.listener)
}
