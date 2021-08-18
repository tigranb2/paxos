package network

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"paxos/logger"
	"paxos/msg"
	"paxos/queue"
	"time"
)

type Proposer struct {
	clientRequest chan *msg.QueueRequest
	connections   map[string]msg.MessengerClient //store server connections for optimization
	data          map[int32]*msg.Msg
	id            int
	ip            string //socket for Proposer server
	queue         queue.PriorityQueue
	quorum        int
	quorumsData   map[int32]*quorumData //keeps track of responses for quorum
	quorumStatus  chan *msg.Msg
}

type quorumData struct {
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(connections []string, id int, ip string, quorum int) Proposer {
	return Proposer{clientRequest: make(chan *msg.QueueRequest), connections: initConnectionsMap(connections), data: make(map[int32]*msg.Msg), id: id, ip: ip, quorum: quorum, quorumsData: make(map[int32]*quorumData), quorumStatus: make(chan *msg.Msg)}
}

func (p *Proposer) Run() {
	var slot int32
	timer := time.After(5 * time.Second) //timer controls how often request are taken from queue
	go serverInit(p)                     //start proposer server
	for {
		select {
		case clientReq := <-p.clientRequest:
			p.queue.Push(*clientReq)
			p.clientRequest <- &msg.QueueRequest{} //ACK
		case <-timer:
			if len(p.queue) != 0 {
				req := p.queue.Pop().(msg.QueueRequest)
				slot++
				p.data[slot] = &msg.Msg{
					Type:      msg.Type_Prepare,
					SlotIndex: slot,
					Id:        time.Now().UnixNano(),
					Value:     req.GetValue(),
					Priority:  req.GetPriority()}
				p.quorumsData[slot] = &quorumData{}
				go p.broadcast(slot) //broadcast new msg
			}
			timer = time.After(5 * time.Second) //restart timer
		case rec := <-p.quorumStatus:
			slotQuorumData := p.quorumsData[rec.GetSlotIndex()]
			proposerMsg := p.data[rec.GetSlotIndex()]
			if rec.GetType() == msg.Type_Prepare {
				if slotQuorumData.promises >= p.quorum { //prepare phase passed
					proposerMsg.Type = msg.Type_Propose
				} else { //prepare phase failed
					slotQuorumData.promises = 0
					proposerMsg.Id = time.Now().UnixNano()
				}
			} else {
				if slotQuorumData.accepts >= p.quorum { //propose phase passed
					file := fmt.Sprintf("proposer-%v.log", p.id)
					str := fmt.Sprintf("Slot: %v, value: %v\n", proposerMsg.GetSlotIndex(), proposerMsg.GetValue())
					logger.WriteToLog(file, str)
					continue
				} else if slotQuorumData.proposeAttempt == 3 { //propose phase failed 3 times
					slotQuorumData.proposeAttempt = 0
					slotQuorumData.promises = 0
					slotQuorumData.accepts = 0
					proposerMsg.Type = msg.Type_Prepare
					proposerMsg.Id = time.Now().UnixNano()
				} else { //propose phase failed less than 3 times
					slotQuorumData.accepts = 0
					slotQuorumData.proposeAttempt++
				}
			}
			go p.broadcast(proposerMsg.GetSlotIndex())
		}
	}
}

func (p *Proposer) broadcast(slot int32) {
	for netIp := range p.connections {
		p.msgServer(netIp, p.data[slot])
	}

	n := rand.Intn(1000)
	time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties

	p.quorumStatus <- p.data[slot]
}

func (p *Proposer) msgServer(ip string, data *msg.Msg) {
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
	rec, err := c.MsgAcceptor(ctx, data)
	if err != nil {
		return
	}
	p.handleResponse(rec, err)
}

func (p *Proposer) handleResponse(rec *msg.Msg, err error) {
	if err != nil { //simulate timeout
		return
	}
	proposerMsg, ok := p.data[rec.GetSlotIndex()]
	if !ok {
		return
	}
	slotQuorumData := p.quorumsData[rec.GetSlotIndex()]
	switch rec.GetType() {
	case msg.Type_Promise:
		if proposerMsg.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
			slotQuorumData.promises++
			//handles case where acceptor accepted previous value
			if rec.GetPreviousId() != 0 && rec.GetPreviousId() > proposerMsg.GetPreviousId() { //checks if previously accepted value is the newest one seen
				if proposerMsg.GetPreviousId() == 0 { //check whether proposerMsg still has original write value
					p.queue.Push(msg.QueueRequest{Priority: proposerMsg.GetPriority(), Value: proposerMsg.GetValue()}) //push original write to head of queue
				}
				proposerMsg.Value = rec.GetValue()
				proposerMsg.PreviousId = rec.GetPreviousId()
			}
		}
	case msg.Type_Accept:
		if proposerMsg.GetType() == msg.Type_Propose {
			slotQuorumData.accepts++
		}
	}
}

func initConnectionsMap(connectionsSlice []string) map[string]msg.MessengerClient {
	connectionsMap := make(map[string]msg.MessengerClient)
	for _, ip := range connectionsSlice {
		connectionsMap[ip] = nil
	}
	return connectionsMap
}
