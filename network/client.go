package network

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"paxos/msg"
	"time"
)

func msgServer(ip string, p *Proposer) {
	//dials server if connection does not already exist
	if p.connections[ip] == nil {
		conn, err := grpc.Dial(ip, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := msg.NewMessengerClient(conn)
		p.connections[ip] = c //save connection to proposer's map
	}

	c := p.connections[ip]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//Send message to proposer with 1 sec timeout
	rec, err := c.SendMsg(ctx, &p.data)
	p.handleResponse(rec, err)
}
