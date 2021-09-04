package roles

import (
	"context"
	"paxos/msg"
	"paxos/randstring"
	"time"
)

type Client struct {
	connections      map[int]msg.MessengerClient //connections to proposers
	connectionStatus chan *msg.SlotValue
	ips              []string
	originalId       int
	proposerId       int
	requests         chan *msg.QueueRequest
	stopUnicast      chan bool
	timeout          int //timeout time in seconds
}

func InitClient(id int, ips []string, timeout int) *Client {
	return &Client{connections: createConnections(ips), connectionStatus: make(chan *msg.SlotValue), ips: ips, originalId: id, proposerId: id, requests: make(chan *msg.QueueRequest), stopUnicast: make(chan bool), timeout: timeout}
}

func (c *Client) CloseLoopClient() {
	go c.unicast()
	for {
		req := &msg.QueueRequest{Priority: time.Now().UnixNano(), Value: randstring.FixedLengthString(10)}
		c.requests <- req
		select {
		case resp := <-c.connectionStatus:
			if resp != nil { //Proposer responded to request with committed value
				continue
			} else { //Connections with proposer failed
				c.stopUnicast <- true

				//client moves on to proposer of adjacent id
				if c.proposerId < len(c.connections) {
					c.proposerId++
				} else {
					c.proposerId = 1
				}

				if c.proposerId == c.originalId { //client terminates once it has timed out from all proposers
					return
				}

				go c.unicast()
			}
		}
	}

}

func (c *Client) unicast() {
	//dials server if connection does not already exist
	if _, ok := c.connections[c.proposerId]; !ok {
		if conn := dialServer(c.ips[c.proposerId-1]); conn != nil {
			c.connections[c.proposerId] = conn //save connection to proposer's map
		} else {
			c.connectionStatus <- nil
		}
	}
	conn := c.connections[c.proposerId]

	stream, err := conn.MsgProposer(context.Background())
	if err != nil {
		c.connectionStatus <- nil
	}

	//receive messages
	go func() {
		for {
			rec, err := stream.Recv()
			if err != nil {
				c.connectionStatus <- nil
			}

			c.connectionStatus <- rec
		}
	}()

	//send messages
	go func() {
		for {
			select {
			case req := <-c.requests:
				if err := stream.Send(req); err != nil {
					c.connectionStatus <- nil
				}
			}
		}
	}()

	<-c.stopUnicast
}
