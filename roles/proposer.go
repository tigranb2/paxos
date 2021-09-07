package roles

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand"
	"paxos/msg"
	"paxos/queue"
	"sync"
	"time"
)

type Proposer struct {
	clientRequest     chan interface{}
	connections       map[int]msg.MessengerClient //store server connections for optimization
	id                int
	ip                string //socket for Proposer server
	int64Needed       int    //number of int64 needed
	ips               []string
	learnerMsgs       chan *msg.SlotValue //messages for learner go here
	ledger            map[int32]string    //stores committed values for slots
	pendingMsg        *msg.Msg
	queue             queue.PriorityQueue
	quorum            int
	quorumsData       *quorumData //keeps track of responses for quorum
	quorumStatus      chan *msg.Msg
	routinesCompleted int //represents how many RPC to Proposer have terminated
}

type quorumData struct {
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(int64Needed int, ips []string, id int, ip string, quorum int) Proposer {
	return Proposer{int64Needed: int64Needed, clientRequest: make(chan interface{}), connections: createConnections(ips), id: id, ip: ip, ips: ips, learnerMsgs: make(chan *msg.SlotValue), ledger: make(map[int32]string), queue: make(queue.PriorityQueue, 0), quorum: quorum, quorumStatus: make(chan *msg.Msg)}
}

var commits float64
var start time.Time

func (p *Proposer) Run() {
	go p.learner()
	var slot int32
	go serverInit(p) //start proposer server
	for {
		select {
		case clientReq := <-p.clientRequest:
			heap.Push(&p.queue, clientReq.(msg.QueueRequest))

			if commits == 0 {
				start = time.Now()
			}
			if p.pendingMsg == nil {
				p.initSlot(&slot)
			}
		case rec := <-p.quorumStatus:
			if _, ok := p.ledger[p.pendingMsg.GetSlotIndex()]; ok { //check whether slot already has committed value
				p.ledger[p.pendingMsg.GetSlotIndex()] = p.pendingMsg.GetValue()
				p.commitValue(p.pendingMsg)
				p.initSlot(&slot)
				continue
			}

			if rec.GetType() == msg.Type_Prepare {
				if p.quorumsData.promises >= p.quorum { //prepare phase passed
					p.pendingMsg.Type = msg.Type_Propose
				} else { //prepare phase failed
					p.quorumsData.promises = 0
					p.pendingMsg.Id = time.Now().UnixNano()
				}
			} else {
				if p.quorumsData.accepts >= p.quorum { //propose phase passed
					p.commitValue(p.pendingMsg)
					p.initSlot(&slot)
					continue
				} else if p.quorumsData.proposeAttempt == 3 { //propose phase failed 3 times
					p.quorumsData.proposeAttempt = 0
					p.quorumsData.promises = 0
					p.quorumsData.accepts = 0
					p.pendingMsg.Type = msg.Type_Prepare
					p.pendingMsg.Id = time.Now().UnixNano()
				} else { //propose phase failed less than 3 times
					p.quorumsData.accepts = 0
					p.quorumsData.proposeAttempt++
				}
			}
			go p.broadcast()
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

func (p *Proposer) broadcast() {
	var wg sync.WaitGroup
	for acceptorId := range p.ips {
		//dials server if connection does not already exist
		if _, ok := p.connections[acceptorId+1]; !ok {
			if c := dialServer(p.ips[acceptorId]); c != nil {
				p.connections[acceptorId+1] = c //save connection to proposer's map
			} else {
				continue
			}
		}
		wg.Add(1)
		go p.callAcceptor(&wg, acceptorId+1, p.pendingMsg)
	}

	wg.Wait()
	n := rand.Intn(1)
	time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties
	p.quorumStatus <- p.pendingMsg
}

func (p *Proposer) callAcceptor(wg *sync.WaitGroup, acceptorId int, data *msg.Msg) {
	defer wg.Done()
	c := p.connections[acceptorId]
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
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

	switch rec.GetType() {
	case msg.Type_Promise:
		if p.pendingMsg.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
			p.quorumsData.promises++
			//handles case where acceptor accepted previous value
			if rec.GetPreviousId() != 0 && rec.GetPreviousId() > p.pendingMsg.GetPreviousId() { //checks if previously accepted value is the newest one seen
				if p.pendingMsg.GetPreviousId() == 0 { //check whether proposerMsg still has original write value
					heap.Push(&p.queue, msg.QueueRequest{Priority: p.pendingMsg.GetPriority(), Value: p.pendingMsg.GetValue()}) //push original write to head of queue
				}
				p.pendingMsg.Value = rec.GetValue()
				p.pendingMsg.PreviousId = rec.GetPreviousId()
			}
		}
	case msg.Type_Accept:
		if p.pendingMsg.GetType() == msg.Type_Propose {
			p.quorumsData.accepts++
		}
	}
}

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	commits++
	p.clientRequest <- &msg.SlotValue{SlotIndex: proposerMsg.GetSlotIndex(), Value: proposerMsg.GetValue()} //ACK
	p.pendingMsg = nil
	p.quorumsData = nil

	if commits == 10000 {
		elapsed := time.Since(start).Seconds()
		fmt.Println(commits / elapsed)
	}
}

func (p *Proposer) initSlot(slot *int32) {
	if len(p.queue) != 0 {
		req := heap.Pop(&p.queue).(msg.QueueRequest)
		*slot++
		p.pendingMsg = &msg.Msg{
			Type:       msg.Type_Prepare,
			SlotIndex:  *slot,
			Id:         time.Now().UnixNano(),
			Value:      req.GetValue(),
			Priority:   req.GetPriority(),
			ProposerId: int32(p.id),
			Size:       make([]int64, p.int64Needed)}
		p.quorumsData = &quorumData{}
		go p.broadcast() //broadcast new msg
	}
}
