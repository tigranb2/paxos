package roles

import (
	"fmt"
	"paxos/msg"
	"sync"
	"time"
)

var commits float64
var start time.Time

type Proposer struct {
	commit          chan int32
	connections     []*TCP //store connections to acceptors
	id              int
	int32Needed     int    //number of int32s needed for designated msg size (each is 4 bytes)
	ip              string //socket for Proposer server
	pendingMsgs     map[int32]*msg.Msg
	quorum          int
	receiveResponse chan interface{}
	wgs             map[int32]*sync.WaitGroup //WaitGroup for each slot broadcasting messages
}

func InitProposer(int32Needed int, ips []string, id int, ip string, quorum int) Proposer { //connections: createConnections(ips),
	receiveResponse := make(chan interface{})
	var connections []*TCP
	for _, acceptorIp := range ips {
		t := initTCP(msg.ToAcceptor, acceptorIp, receiveResponse)
		connections = append(connections, t)

		go t.sendMsgs()
		go t.receiveMsgs()
	}

	return Proposer{commit: make(chan int32), connections: connections, id: id, int32Needed: int32Needed, ip: ip, pendingMsgs: make(map[int32]*msg.Msg), quorum: quorum, receiveResponse: receiveResponse, wgs: make(map[int32]*sync.WaitGroup)}
}

func (p *Proposer) Run() {
	timerDuration := time.Microsecond * 10
	timer := time.After(timerDuration) //timer controls how often request are taken from queue

	var slot int32
	for {
		select {
		case <-timer:
			if slot == 0 {
				start = time.Now()
			}

			slot++
			p.initSlot(slot)

			p.wgs[slot].Add(len(p.connections))
			go p.broadcast(p.wgs[slot], p.pendingMsgs[slot]) //broadcast new msg

			timer = time.After(timerDuration) //restart timer
		case resp := <-p.receiveResponse:
			r := resp.(*msg.Msg)
			p.wgs[r.GetSlotIndex()].Add(-1) //decrement WaitGroup for broadcast
		case s := <-p.commit:
			p.commitValue(p.pendingMsgs[s])
		}
	}
}

func (p *Proposer) broadcast(wg *sync.WaitGroup, msg *msg.Msg) {
	for i := 0; i < len(p.connections); i++ {
		p.connections[i].sendChan <- msg
	}

	wg.Wait()
	p.commit <- msg.GetSlotIndex()
}

func (p *Proposer) initSlot(slot int32) {
	p.pendingMsgs[slot] = &msg.Msg{
		Type:       msg.Prepare,
		SlotIndex:  slot,
		Id:         time.Now().UnixNano(),
		Value:      "req.GetValue()",
		ProposerId: int32(p.id),
		Size_:      make([]int32, p.int32Needed)}
	var wg sync.WaitGroup
	p.wgs[slot] = &wg
}

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	delete(p.pendingMsgs, proposerMsg.GetSlotIndex())

	commits++
	fmt.Println("commited", commits)
	if commits == 10000 {
		elapsed := time.Since(start).Seconds()
		fmt.Println(commits / elapsed)
	}
}
