package roles

import (
	"container/heap"
	"fmt"
	"paxos/msg"
	"paxos/queue"
	"time"
)

var commits float64
var start time.Time

type Proposer struct {
	clientRequest    chan interface{}
	clientWriteChans map[int32]chan *msg.SlotValue //channels used for writing responses to clients
	connections      []*TCP                        //store connections to acceptors
	id               int
	int64Needed      int           //number of int64s needed
	ip               string        //socket for Proposer server
	learnerMsgs      chan *msg.Msg //messages for learner go here
	msgs             map[int32]*msg.Msg
	queue            queue.PriorityQueue
	quorum           int
	receiveResponse  chan interface{}
	slotData         map[int32]*quorumData //keeps track of responses for quorum
}

type quorumData struct {
	responses         int
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(int64Needed int, ips []string, id int, ip string, quorum int) Proposer { //connections: createConnections(ips),
	var connections []*TCP
	receiveResponse := make(chan interface{})
	for _, acceptorIp := range ips {
		t := initTCP(msg.SendingMsg, acceptorIp, receiveResponse)
		connections = append(connections, t)

		go t.sendMsgs()
		go t.receiveMsgs()
		t.sendChan <- &msg.Msg{ProposerId: int32(id)}
	}

	return Proposer{clientRequest: make(chan interface{}), clientWriteChans: make(map[int32]chan *msg.SlotValue), connections: connections, id: id, int64Needed: int64Needed, ip: ip, learnerMsgs: make(chan *msg.Msg), msgs: make(map[int32]*msg.Msg), queue: make(queue.PriorityQueue, 0), quorum: quorum, slotData: make(map[int32]*quorumData), receiveResponse: receiveResponse}
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
				p.broadcast(slot) //broadcast new msg
			}
			timer = time.After(timerDuration) //restart timer
		case resp := <-p.receiveResponse:
			r := resp.(*msg.Msg)

			//send message to learner if it is designated as a LearnerMsg
			switch r.GetType() {
			case msg.LearnerMsg:
				p.learnerMsgs <- r
				continue
			case msg.Commit:
				p.msgs[r.GetSlotIndex()] = r
			default:
				break
			}

			p.handleResponse(r)
			p.slotData[r.GetSlotIndex()].responses++
			p.checkQuorum(r.GetSlotIndex())
		}
	}
}

func (p *Proposer) learner() { //manages proposer's ledger
	type msgMin struct { //minimized version of msg.SlotValue for learner's acceptsPerSlot map
		slotIndex int32
		value     string
	}

	acceptsPerSlot := make(map[msgMin]int)
	for {
		select {
		case rec := <-p.learnerMsgs:
			m := msgMin{rec.GetSlotIndex(), rec.GetValue()}
			acceptsPerSlot[m]++

			if acceptsPerSlot[m] == p.quorum {
				p.receiveResponse <- &msg.Msg{Type: msg.Commit, SlotIndex: rec.GetSlotIndex(), Value: rec.GetValue()}
			}
		}
	}
}

func (p *Proposer) broadcast(slot int32) {
	for i := 0; i < len(p.connections); i++ {
		p.connections[i].sendChan <- p.msgs[slot]
	}
}

func (p *Proposer) handleResponse(rec *msg.Msg) {
	if rec == nil { //simulate timeout
		return
	}
	proposerMsg, ok := p.msgs[rec.GetSlotIndex()]
	if !ok {
		return
	}
	slotQuorumData := p.slotData[rec.GetSlotIndex()]
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
	p.msgs[slot] = &msg.Msg{
		Id:         time.Now().UnixNano(),
		Priority:   req.GetPriority(),
		Type:       msg.Prepare,
		SlotIndex:  slot,
		ProposerId: int32(p.id),
		FromClient: req.GetFromClient(),
		Value:      req.GetValue(),
		Size_:      make([]int64, p.int64Needed)}
	p.slotData[slot] = &quorumData{}
}

func (p *Proposer) checkQuorum(slot int32) {
	sData := p.slotData[slot]
	proposerMsg := p.msgs[slot]

	//check whether slot already has committed value
	if proposerMsg.GetType() == msg.Commit {
		p.commitValue(proposerMsg)
		return
	}

	if proposerMsg.GetType() == msg.Prepare {
		if sData.promises >= p.quorum { //prepare phase passed
			proposerMsg.Type = msg.Propose
		} else { //prepare phase not passed
			//phase will only fail if all acceptors have responded (and quorum not met)
			if sData.responses < len(p.connections) {
				return
			}

			sData.promises = 0
			proposerMsg.Id = time.Now().UnixNano()
		}
	} else {
		if sData.accepts >= p.quorum { //propose phase passed
			p.msgs[proposerMsg.GetSlotIndex()] = &msg.Msg{Type: msg.Commit, SlotIndex: proposerMsg.GetSlotIndex(), Value: proposerMsg.GetValue()}
			p.commitValue(proposerMsg)
			return
		} else if sData.proposeAttempt == 3 { //propose phase failed 3 times
			sData.proposeAttempt = 0
			sData.promises = 0
			sData.accepts = 0
			proposerMsg.Type = msg.Prepare
			proposerMsg.Id = time.Now().UnixNano()
		} else { //propose phase failed less than 3 times
			//phase will only fail if all acceptors have responded (and quorum not met)
			if sData.responses < len(p.connections) {
				return
			}

			sData.accepts = 0
			sData.proposeAttempt++
		}
	}
	sData.responses = 0
	p.broadcast(proposerMsg.GetSlotIndex())
}

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	commits++
	p.clientWriteChans[proposerMsg.GetFromClient()] <- &msg.SlotValue{SlotIndex: proposerMsg.GetSlotIndex(), Value: proposerMsg.GetValue()} //ACK
	delete(p.slotData, proposerMsg.GetSlotIndex())

	if commits == 10000 {
		elapsed := time.Since(start).Seconds()
		fmt.Println(commits / elapsed)
	}
}
