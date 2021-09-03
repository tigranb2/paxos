package roles

import (
	"context"
	"paxos/msg"
	"paxos/randstring"
	"time"
)

type Client struct {
	ack         chan *msg.SlotValue
	connections map[int]msg.MessengerClient //connections to proposers
	ips         []string
	originalId  int
	proposerId  int
	timeout     int //timeout time in seconds
}

func InitClient(id int, ips []string, timeout int) *Client {
	return &Client{ack: make(chan *msg.SlotValue), connections: createConnections(ips), ips: ips, originalId: id, proposerId: id, timeout: timeout}
}

func (c *Client) CloseLoopClient() {
	for {
		req := &msg.QueueRequest{Priority: time.Now().UnixNano(), Value: randstring.FixedLengthString(10)}
		go c.unicast(req)
		select {
		case resp := <-c.ack:
			if resp != nil { //Proposer acknowledged request
				continue
			} else { //Proposer did not acknowledge request
				//client moves on to proposer of adjacent id
				if c.proposerId < len(c.connections) {
					c.proposerId++
				} else {
					c.proposerId = 1
				}

				if c.proposerId == c.originalId { //client terminates once it has timed out from all proposers
					return
				}
			}
		}
	}

}

func (client *Client) unicast(req *msg.QueueRequest) {
	//dials server if connection does not already exist
	if _, ok := client.connections[client.proposerId]; !ok {
		if c := dialServer(client.ips[client.proposerId-1]); c != nil {
			client.connections[client.proposerId] = c //save connection to proposer's map
		} else {
			return
		}
	}
	conn := client.connections[client.proposerId]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(client.timeout)*time.Second)
	defer cancel()
	resp, err := conn.MsgProposer(ctx, req)
	if err == nil {
		client.ack <- resp
	} else {
		client.ack <- nil
	}
}
