package network

import (
	"math/rand"
	"paxos/msg"
	"time"
)

type Proposer struct {
	data              msg.Msg                        //stores proposed id, value, and current phase
	promises, accepts int                            //keeps track of responses for quorum
	connections       map[string]msg.MessengerClient //store server connections for optimization
	quorum            int
	quorumStatus      chan bool
}

func InitProposer(value string, connections []string, quorum int) Proposer {
	return Proposer{data: msg.Msg{Type: msg.Type_Prepare, Id: time.Now().UnixNano(), Value: value}, connections: initConnectionsMap(connections), quorum: quorum, quorumStatus: make(chan bool)}
}

func (p *Proposer) Run(stop chan string) {
	proposeAttempt := 0
	for {
		select {
		case quorumMet := <-p.quorumStatus:
			if quorumMet == true {
				if p.data.GetType() == msg.Type_Prepare {
					//move from prepare to propose phase
					p.data.Type = msg.Type_Propose
					p.Broadcast()
				} else {
					//move from propose phase to termination (consensus reached)
					stop <- p.data.GetValue()
				}
			} else {
				p.promises, p.accepts = 0, 0
				if p.data.GetType() == msg.Type_Prepare {
					//proposer restarts prepare phase with a different id
					p.data.Id = time.Now().UnixNano()
				} else if proposeAttempt == 3 {
					//proposer returns to prepare phase after 3 failed propose attempts
					proposeAttempt = 0
					p.data.Type = msg.Type_Prepare
					p.data.Id = time.Now().UnixNano()
				} else if p.data.GetType() == msg.Type_Propose {
					proposeAttempt++
				}
				p.Broadcast()
			}
		}
	}
}

func (p *Proposer) Broadcast() {
	for netIp := range p.connections {
		msgServer(netIp, p)
	}

	n := rand.Intn(1000)
	time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties

	if p.data.GetType() == msg.Type_Prepare && p.promises < p.quorum { //prepare phase fails
		p.quorumStatus <- false
	} else if p.data.GetType() == msg.Type_Prepare && p.promises >= p.quorum { //prepare phase succeeds
		p.quorumStatus <- true
	} else if p.data.GetType() == msg.Type_Propose && p.accepts < p.quorum { //propose phase fails
		p.quorumStatus <- false
	} else if p.data.GetType() == msg.Type_Propose && p.accepts >= p.quorum { //propose phase succeeds
		p.quorumStatus <- true
	}
}

func (p *Proposer) handleResponse(rec *msg.Msg, err error) {
	if err != nil { //simulate timeout
		return
	}
	switch rec.GetType() {
	case msg.Type_Promise:
		if p.data.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
			p.promises++
			//handles case where acceptor accepted previous value
			if rec.GetPreviousId() != 0 && rec.GetPreviousId() > p.data.GetPreviousId() { //checks if previously accepted value is newest one seen
				p.data.Value = rec.GetValue()
				p.data.PreviousId = rec.GetPreviousId()
			}
		}
	case msg.Type_Accept:
		if p.data.GetType() == msg.Type_Propose {
			p.accepts++
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
