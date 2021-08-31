package roles

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"paxos/msg"
	"time"
)

type server struct {
	msg.UnimplementedMessengerServer
	node interface{}
}

func (s *server) MsgAcceptor(ctx context.Context, msg *msg.Msg) (*msg.Msg, error) {
	a, ok := s.node.(*Acceptor)
	if !ok {
		log.Fatalf("destination server must be an acceptor")
	}
	return a.acceptMsg(msg)
}

func (s *server) MsgProposer(ctx context.Context, request *msg.QueueRequest) (*msg.SlotValue, error) {
	p, ok := s.node.(*Proposer)
	if !ok {
		log.Fatalf("destination server must be a proposer")
	}

	p.clientRequest <- *request
	for {
		select {
		case resp := <-p.clientRequest:
			if _, ok := resp.(msg.QueueRequest); ok { //do not read other Client's requests
				continue
			}
			return resp.(*msg.SlotValue), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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

func (s *server) MsgLearner(ctx context.Context, data *msg.SlotValue) (*msg.Empty, error) {
	p, ok := s.node.(*Proposer)
	if !ok {
		log.Fatalf("destination server must be a proposer")
	}

	p.learnerMsgs <- data
	return &msg.Empty{}, nil
}

func createConnections(ips []string) map[int]msg.MessengerClient {
	connections := make(map[int]msg.MessengerClient)
	for id, ip := range ips {
		if c := dialServer(ip); c != nil {
			connections[id+1] = c
		}
	}
	return connections
}

func dialServer(ip string) msg.MessengerClient {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(ctx, ip, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil
	}

	return msg.NewMessengerClient(conn)
}
