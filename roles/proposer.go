package roles

import (
	"fmt"
	"paxos/msg"
	"paxos/network"
	"time"
)

type Proposer struct {
	Data           msg.Msg //stores proposed id, value, and current phase
	nodesResponded int     //keeps track of responses for quorum
	Connections map[string]msg.MessengerClient //store server connections for optimization
}

func InitProposer(value string) Proposer {
	return Proposer{Data: msg.Msg{Type: msg.Type_Prepare, Value: value}, Connections: make(map[string]msg.MessengerClient)}
}

func (p *Proposer) Run(quorum int) {
	for {
		if p.nodesResponded < quorum { //if proposer failed to get enough promises/accepts
			p.Data.Type = msg.Type_Prepare
			p.Data.Id = time.Now().UnixNano()
		} else if p.Data.GetType() == msg.Type_Prepare { //when proposer has enough promises
			p.Data.Type = msg.Type_Propose
		} else { //when consensus is reached
			break
		}
		p.nodesResponded = 0

		for netIp, _ := range p.Connections {
			network.MsgServer(netIp, p)
		}
	}

	fmt.Printf("Consensus reached! Value: %v\n", p.Data.Value)
}

func (p *Proposer) HandleResponse(rec *msg.Msg, err error) {
	if err != nil { //simulate timeout
		return
	}
	switch rec.GetType() {
	case msg.Type_Promise:
		if p.Data.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
			p.nodesResponded++
			//handles case where acceptor accepted previous value
			if rec.GetPreviousId() != 0 && rec.GetPreviousId() > p.Data.GetPreviousId(){ //checks if previously accepted value is newest one seen
				p.Data.Value = rec.GetValue()
				p.Data.PreviousId = rec.GetPreviousId()
			}
		}
	case msg.Type_Accept:
		if p.Data.GetType() == msg.Type_Propose {
			p.nodesResponded++
		}
	}
}
