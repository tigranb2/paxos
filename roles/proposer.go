package roles

import (
	"container/heap"
	"context"
	"math/rand"
	"paxos/msg"
	"paxos/queue"
	"time"
)

type Proposer struct {
	clientRequest chan interface{}
	connections   map[int]msg.MessengerClient //store server connections for optimization
	id            int
	ip            string //socket for Proposer server
	ips           []string
	learnerMsgs   chan *msg.SlotValue //messages for learner go here
	ledger        map[int32]string    //stores committed values for slots
	pendingMsgs   map[int32]*msg.Msg
	queue         queue.PriorityQueue
	quorum        int
	quorumsData   map[int32]*quorumData //keeps track of responses for quorum
	quorumStatus  chan *msg.Msg
}

type quorumData struct {
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(ips []string, id int, ip string, quorum int) Proposer {
	return Proposer{clientRequest: make(chan interface{}), connections: createConnections(ips), id: id, ip: ip, ips: ips, learnerMsgs: make(chan *msg.SlotValue), ledger: make(map[int32]string), pendingMsgs: make(map[int32]*msg.Msg), queue: make(queue.PriorityQueue, 0), quorum: quorum, quorumsData: make(map[int32]*quorumData), quorumStatus: make(chan *msg.Msg)}
}

func (p *Proposer) Run() {
	go p.learner()
	var slot int32
	go serverInit(p) //start proposer server
	for {
		select {
		case clientReq := <-p.clientRequest:
			heap.Push(&p.queue, clientReq.(msg.QueueRequest))
			if len(p.pendingMsgs) == 0 {
				p.initSlot(&slot)
			}
		case rec := <-p.quorumStatus:
			slotQuorumData := p.quorumsData[rec.GetSlotIndex()]
			proposerMsg := p.pendingMsgs[rec.GetSlotIndex()]

			if _, ok := p.ledger[proposerMsg.GetSlotIndex()]; ok { //check whether slot already has committed value
				p.ledger[proposerMsg.GetSlotIndex()] = proposerMsg.GetValue()
				p.commitValue(proposerMsg)
				p.initSlot(&slot)
				continue
			}

			if rec.GetType() == msg.Type_Prepare {
				if slotQuorumData.promises >= p.quorum { //prepare phase passed
					proposerMsg.Type = msg.Type_Propose
				} else { //prepare phase failed
					slotQuorumData.promises = 0
					proposerMsg.Id = time.Now().UnixNano()
				}
			} else {
				if slotQuorumData.accepts >= p.quorum { //propose phase passed
					p.commitValue(proposerMsg)
					p.initSlot(&slot)
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

func (p *Proposer) learner() { //manages proposer's ledger
	type slotValueMin struct { //minimized version of msg.SlotValue for learner's acceptsPerSlot map
		slotIndex int32
		value     string
	}

	acceptsPerSlot := make(map[slotValueMin]int)
	for {
		select {
		case rec := <-p.learnerMsgs:
			if rec.GetType() == msg.Type_Accept {
				m := slotValueMin{rec.GetSlotIndex(), rec.GetValue()}
				acceptsPerSlot[m]++

				if acceptsPerSlot[m] == p.quorum {
					p.ledger[m.slotIndex] = m.value
				}
			} else { //if slot has already been committed
				p.ledger[rec.GetSlotIndex()] = rec.GetValue()
			}
		}
	}
}

func (p *Proposer) broadcast(slot int32) {
	for acceptorId := range p.ips {
		//dials server if connection does not already exist
		if _, ok := p.connections[acceptorId+1]; !ok {
			if c := dialServer(p.ips[acceptorId]); c != nil {
				p.connections[acceptorId+1] = c //save connection to proposer's map
			} else {
				continue
			}
		}
		p.callAcceptor(acceptorId+1, p.pendingMsgs[slot])
	}

	n := rand.Intn(1000)
	time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties
	p.quorumStatus <- p.pendingMsgs[slot]
}

func (p *Proposer) callAcceptor(acceptorId int, data *msg.Msg) {
	c := p.connections[acceptorId]
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
	proposerMsg, ok := p.pendingMsgs[rec.GetSlotIndex()]
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
					heap.Push(&p.queue, msg.QueueRequest{Priority: proposerMsg.GetPriority(), Value: proposerMsg.GetValue()}) //push original write to head of queue
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

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	p.clientRequest <- &msg.SlotValue{SlotIndex: proposerMsg.GetSlotIndex(), Value: proposerMsg.GetValue()} //ACK
	delete(p.pendingMsgs, proposerMsg.GetSlotIndex())
	delete(p.quorumsData, proposerMsg.GetSlotIndex())
}

func (p *Proposer) initSlot(slot *int32) {
	if len(p.queue) != 0 {
		req := heap.Pop(&p.queue).(msg.QueueRequest)
		*slot++
		p.pendingMsgs[*slot] = &msg.Msg{
			Type:       msg.Type_Prepare,
			SlotIndex:  *slot,
			Id:         time.Now().UnixNano(),
			Value:      req.GetValue(),
			Priority:   req.GetPriority(),
			ProposerId: int32(p.id)}
		p.quorumsData[*slot] = &quorumData{}
		go p.broadcast(*slot) //broadcast new msg
	}
}
