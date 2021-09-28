package roles

import (
	"paxos/msg"
	"paxos/randstring"
	"time"
)

type Client struct {
	connections     map[int]*TCP //connections to proposers
	ips             []string
	originalId      int
	proposerId      int
	receiveResponse chan interface{}
	requests        chan *msg.QueueRequest
	stopUnicast     chan bool
}

func InitClient(id int, ips []string) *Client {
	receiveResponse := make(chan interface{})
	connections := make(map[int]*TCP)
	for proposerId, proposerIp := range ips {
		t := initTCP(msg.SendingRequest, proposerIp, receiveResponse)
		connections[proposerId+1] = t
	}

	return &Client{connections: connections, receiveResponse: receiveResponse, ips: ips, originalId: id, proposerId: id, requests: make(chan *msg.QueueRequest), stopUnicast: make(chan bool)}
}

func (c *Client) CloseLoopClient() {
	go c.connections[c.proposerId].sendMsgs()
	go c.connections[c.proposerId].receiveMsgs()

	for {
		req := &msg.QueueRequest{Priority: time.Now().UnixNano(), Value: randstring.FixedLengthString(10)}
		c.connections[c.proposerId].sendChan <- req
		select {
		case <-c.receiveResponse:
			continue
		}
	}

}
