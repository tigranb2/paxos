package roles

import (
	"container/heap"
	"fmt"
	"paxos/msg"
	"paxos/queue"
	"sync"
	"time"
)

var commits float64
var start time.Time

type Proposer struct {
	clientRequest    chan interface{}
	clientWriteChans map[int32]chan *msg.SlotValue //channels used for writing responses to clients
	connections      map[int]*TCP                  //store connections to acceptors
	id               int
	int64Needed      int                 //number of int64s needed
	ip               string              //socket for Proposer server
	learnerMsgs      chan *msg.SlotValue //messages for learner go here
	ledger           map[int32]string    //stores committed values for slots
	pendingMsgs      map[int32]*msg.Msg
	queue            queue.PriorityQueue
	quorum           int
	quorumsData      map[int32]*quorumData //keeps track of responses for quorum
	quorumStatus     chan *msg.Msg
	receiveResponse  chan interface{}
	wgs              map[int32]*sync.WaitGroup //WaitGroup for each slot broadcasting messages
}

type quorumData struct {
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(int64Needed int, ips []string, id int, ip string, quorum int) Proposer { //connections: createConnections(ips),
	receiveResponse := make(chan interface{})
	connections := make(map[int]*TCP)
	for acceptorId, acceptorIp := range ips {
		t := initTCP(msg.ToAcceptor, acceptorIp, receiveResponse)
		connections[acceptorId+1] = t

		go t.sendMsgs()
		go t.receiveMsgs()
	}

	return Proposer{clientRequest: make(chan interface{}), clientWriteChans: make(map[int32]chan *msg.SlotValue), connections: connections, id: id, int64Needed: int64Needed, ip: ip, learnerMsgs: make(chan *msg.SlotValue), ledger: make(map[int32]string), pendingMsgs: make(map[int32]*msg.Msg), queue: make(queue.PriorityQueue, 0), quorum: quorum, quorumsData: make(map[int32]*quorumData), quorumStatus: make(chan *msg.Msg), receiveResponse: receiveResponse, wgs: make(map[int32]*sync.WaitGroup)}
}

func (p *Proposer) Run() {
	go p.proposerServer()
	go p.learner()
	timerDuration := time.Microsecond * 10
	timer := time.After(timerDuration) //timer controls how often request are taken from queue

	var slot int32
	for {
		select {
		case clientReq := <-p.clientRequest:
			heap.Push(&p.queue, *clientReq.(*msg.QueueRequest))
		case <-timer:
			if len(p.queue) != 0 {
				if slot == 0 {
					start = time.Now()
				}

				slot++
				req := heap.Pop(&p.queue).(msg.QueueRequest)
				p.initSlot(slot, req)
				go p.broadcast(slot) //broadcast new msg
			}
			timer = time.After(timerDuration) //restart timer
		case resp := <-p.receiveResponse:
			r := resp.(*msg.Msg)
			p.handleResponse(r)
			p.wgs[r.GetSlotIndex()].Add(-1) //decrement WaitGroup for broadcast
		case rec := <-p.quorumStatus:
			slotQuorumData := p.quorumsData[rec.GetSlotIndex()]
			proposerMsg := p.pendingMsgs[rec.GetSlotIndex()]

			if _, ok := p.ledger[proposerMsg.GetSlotIndex()]; ok { //check whether slot already has committed value
				p.ledger[proposerMsg.GetSlotIndex()] = proposerMsg.GetValue()
				p.commitValue(proposerMsg)
				continue
			}

			if rec.GetType() == msg.Prepare {
				if slotQuorumData.promises >= p.quorum { //prepare phase passed
					proposerMsg.Type = msg.Propose
				} else { //prepare phase failed
					slotQuorumData.promises = 0
					proposerMsg.Id = time.Now().UnixNano()
				}
			} else {
				if slotQuorumData.accepts >= p.quorum { //propose phase passed
					p.commitValue(proposerMsg)
					continue
				} else if slotQuorumData.proposeAttempt == 3 { //propose phase failed 3 times
					slotQuorumData.proposeAttempt = 0
					slotQuorumData.promises = 0
					slotQuorumData.accepts = 0
					proposerMsg.Type = msg.Prepare
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
			if rec.GetType() == msg.Accept {
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
	var wg sync.WaitGroup
	p.wgs[slot] = &wg

	for i := 1; i <= len(p.connections); i++ {
		p.wgs[slot].Add(1)
		p.connections[i].sendChan <- p.pendingMsgs[slot]
	}

	wg.Wait()
	//n := rand.Intn(1)
	//time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties
	p.quorumStatus <- p.pendingMsgs[slot]
}

func (p *Proposer) handleResponse(rec *msg.Msg) {
	if rec == nil { //simulate timeout
		return
	}
	proposerMsg, ok := p.pendingMsgs[rec.GetSlotIndex()]
	if !ok {
		return
	}
	slotQuorumData := p.quorumsData[rec.GetSlotIndex()]
	switch rec.GetType() {
	case msg.Promise:
		if proposerMsg.GetType() == msg.Prepare { //check that proposer is asking for promises
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
	case msg.Accept:
		if proposerMsg.GetType() == msg.Propose {
			slotQuorumData.accepts++
		}
	}
}

func (p *Proposer) initSlot(slot int32, req msg.QueueRequest) {
	p.pendingMsgs[slot] = &msg.Msg{
		Type:       msg.Prepare,
		SlotIndex:  slot,
		Id:         time.Now().UnixNano(),
		Value:      req.GetValue(),
		Priority:   req.GetPriority(),
		ProposerId: int32(p.id),
		FromClient: req.GetFromClient()}
	p.quorumsData[slot] = &quorumData{}
}

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	commits++
	p.clientWriteChans[proposerMsg.GetFromClient()] <- &msg.SlotValue{SlotIndex: proposerMsg.GetSlotIndex(), Value: proposerMsg.GetValue()} //ACK
	delete(p.pendingMsgs, proposerMsg.GetSlotIndex())
	delete(p.quorumsData, proposerMsg.GetSlotIndex())

	if commits == 10000 {
		elapsed := time.Since(start).Seconds()
		fmt.Println(commits / elapsed)
	}
}
