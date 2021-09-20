package roles

import (
	"fmt"
	"paxos/msg"
	"time"
)

var commits float64
var start time.Time

type Proposer struct {
	connections     []*TCP //store connections to acceptors
	id              int
	int32Needed     int    //number of int32s needed for designated msg size (each is 4 bytes)
	ip              string //socket for Proposer server
	pendingMsgs     map[int32]*msg.Msg
	quorum          int
	receiveResponse chan interface{}
	slotResp        map[int32]int
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

	return Proposer{connections: connections, id: id, int32Needed: int32Needed, ip: ip, pendingMsgs: make(map[int32]*msg.Msg), quorum: quorum, receiveResponse: receiveResponse, slotResp: make(map[int32]int)}
}

func (p *Proposer) Run() {
	timerDuration := time.Microsecond * 100
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
			for i := 0; i < len(p.connections); i++ {
				p.connections[i].sendChan <- p.pendingMsgs[slot]
			}

			timer = time.After(timerDuration) //restart timer
		case resp := <-p.receiveResponse:
			r := resp.(*msg.Msg)
			p.slotResp[r.GetSlotIndex()]++

			if p.slotResp[r.GetSlotIndex()] == 2 {
				p.commitValue(p.pendingMsgs[r.GetSlotIndex()])
			}
		}
	}
}

func (p *Proposer) initSlot(slot int32) {
	p.pendingMsgs[slot] = &msg.Msg{
		SlotIndex: slot}
}

func (p *Proposer) commitValue(proposerMsg *msg.Msg) {
	delete(p.pendingMsgs, proposerMsg.GetSlotIndex())
	//fmt.Println("committing ", commits)
	commits++
	if commits == 20000 {
		elapsed := time.Since(start).Seconds()
		fmt.Println(commits / elapsed)
	}
}
