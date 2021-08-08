package network

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"paxos/msg"
	"paxos/roles"
)

func MsgServer(ip string, p *roles.Proposer) {
	//dials server if connection does not already exist
	if _, ok := p.Connections[ip]; !ok {
		conn, err := grpc.Dial(ip, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := msg.NewMessengerClient(conn)
		p.Connections[ip] = c //save connection to proposer's map
	}

	c := p.Connections[ip]
	//Send message to proposer
	rec, err := c.SendMsg(context.Background(), &p.Data)
	p.HandleResponse(rec, err)
}