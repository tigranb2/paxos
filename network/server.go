package network

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"paxos/msg"
)

type server struct {
	msg.UnimplementedMessengerServer
	acceptor *Acceptor
}

func (s *server) SendMsg(ctx context.Context, msg *msg.MsgSlice) (*msg.MsgSlice, error) {
	fmt.Printf("Received message %+v\n", msg)
	return s.acceptor.acceptMsg(msg), nil
}

func serverInit(a *Acceptor) {
	ln, err := net.Listen("tcp", a.Ip)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	s := server{acceptor: a}
	grpcServer := grpc.NewServer()
	msg.RegisterMessengerServer(grpcServer, &s)
	if err := grpcServer.Serve(ln); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
