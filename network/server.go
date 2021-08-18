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
	node interface{}
}

func (s *server) MsgAcceptor(ctx context.Context, msg *msg.Msg) (*msg.Msg, error) {
	acceptor, ok := s.node.(*Acceptor)
	if !ok {
		log.Fatalf("destination server must be an acceptor")
	}

	fmt.Printf("Received message %+v\n", msg)
	return acceptor.acceptMsg(msg)
}

func (s *server) MsgProposer(ctx context.Context, request *msg.QueueRequest) (*msg.RequestResponse, error) {
	proposer, ok := s.node.(*Proposer)
	if !ok {
		log.Fatalf("destination server must be a proposer")
	}

	proposer.clientRequest <- request
	select {
	case <-proposer.clientRequest:
		return &msg.RequestResponse{RequestReceived: true}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func serverInit(node interface{}) {
	var ln net.Listener
	var err error
	var s server

	switch t := node.(type) {
	case *Acceptor:
		s = server{node: t}
		ln, err = net.Listen("tcp", t.ip)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	case *Proposer:
		s = server{node: t}
		ln, err = net.Listen("tcp", t.ip)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}

	grpcServer := grpc.NewServer()
	msg.RegisterMessengerServer(grpcServer, &s)
	if err := grpcServer.Serve(ln); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
