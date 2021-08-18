package network

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"paxos/msg"
	"paxos/randstring"
	"time"
)

type Client struct {
	ack              chan bool
	connection       msg.MessengerClient
	connectionSocket string
	timeout          int //timeout time in seconds
}

func InitClient(ip string, timeout int) *Client {
	return &Client{ack: make(chan bool), connectionSocket: ip, timeout: timeout}
}

func (c *Client) CloseLoopClient() {
	for {
		time.Sleep(time.Duration(c.timeout) * time.Second)
		req := &msg.QueueRequest{Priority: time.Now().UnixNano(), Value: randstring.FixedLengthString(10)}
		go c.unicast(req)
		select {
		case status := <-c.ack:
			if status { //Proposer acknowledged request
				continue
			} else { //Proposer did not acknowledge request
				return
			}
		}
	}

}

func (c *Client) unicast(req *msg.QueueRequest) {
	//dials server if connection does not already exist
	if c.connection == nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := grpc.DialContext(ctx, c.connectionSocket, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("unable to connect to proposer on socket %v", c.connectionSocket)
		}

		connection := msg.NewMessengerClient(conn)
		c.connection = connection
	}
	conn := c.connection

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()
	_, err := conn.MsgProposer(ctx, req)
	if err == nil {
		c.ack <- true
	} else {
		c.ack <- false
	}
}
