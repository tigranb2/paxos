package roles

import (
	"fmt"
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
	t := initTCP(msg.SendingRequest, ips[id-1], receiveResponse)
	connections[id] = t

	return &Client{connections: connections, receiveResponse: receiveResponse, ips: ips, originalId: id, proposerId: id, requests: make(chan *msg.QueueRequest), stopUnicast: make(chan bool)}
}

func (c *Client) CloseLoopClient() {
	var latencies int64
	responsesRec := 0
	go c.connections[c.proposerId].sendMsgs()
	go c.connections[c.proposerId].receiveMsgs()

	for {
		req := &msg.QueueRequest{Priority: time.Now().UnixNano(), Value: randstring.FixedLengthString(10)}
		c.connections[c.proposerId].sendChan <- req
		sent := time.Now()
		select {
		case <-c.receiveResponse:
			responsesRec++
			latency := time.Since(sent).Nanoseconds()
			latencies += latency
			if responsesRec == 1000 {
				fmt.Println(float64(latencies/int64(responsesRec)) / 1000000)
			}
		}
	}

}
