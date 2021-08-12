package network

import (
	"context"
	"google.golang.org/grpc"
	"paxos/msg"
	"time"
)

func msgServer(ip string, p *Proposer) {
	//dials server if connection does not already exist
	if p.connections[ip] == nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := grpc.DialContext(ctx, ip, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return
		}

		c := msg.NewMessengerClient(conn)
		p.connections[ip] = c //save connection to proposer's map
	}
	c := p.connections[ip]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//Send message to proposer with 1 sec timeout
	rec, _ := c.SendMsg(ctx, p.data)
	p.handleResponse(rec)
}
