package network

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"paxos/msg"
	"paxos/roles"
)

type server struct {
	msg.UnimplementedMessengerServer
	acceptor *roles.Acceptor
}

func (s *server) SendMsg(ctx context.Context, msg *msg.Msg) (*msg.Msg, error) {
	return s.acceptor.AcceptMsg(msg)
}

func ServerInit(a *roles.Acceptor) {
	ln, err := net.Listen("tcp", a.Ip)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	s := server{acceptor: a}

	grpcServer := grpc.NewServer()
	if err := grpcServer.Serve(ln); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}

	msg.RegisterMessengerServer(grpcServer, &s)
}