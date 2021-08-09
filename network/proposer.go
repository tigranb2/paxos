package network

import (
	"fmt"
	"paxos/msg"
	"time"
)

type Proposer struct {
	data           msg.Msg                        //stores proposed id, value, and current phase
	nodesResponded int                            //keeps track of responses for quorum
	connections    map[string]msg.MessengerClient //store server connections for optimization
}

func InitProposer(value string, connections []string) Proposer {
	return Proposer{data: msg.Msg{Type: msg.Type_Prepare, Value: value}, connections: initConnectionsMap(connections)}
}

func (p *Proposer) Run(quorum int) {
	for {
		if p.nodesResponded < quorum { //if proposer failed to get enough promises/accepts
			p.data.Type = msg.Type_Prepare
			p.data.Id = time.Now().UnixNano()
		} else if p.data.GetType() == msg.Type_Prepare { //when proposer has enough promises
			p.data.Type = msg.Type_Propose
		} else { //when consensus is reached
			break
		}
		p.nodesResponded = 0

		for netIp := range p.connections {
			msgServer(netIp, p)
		}
	}

	fmt.Printf("Consensus reached! Value: %v\n", p.data.Value)
}

func (p *Proposer) handleResponse(rec *msg.Msg, err error) {
	if err != nil { //simulate timeout
		return
	}
	switch rec.GetType() {
	case msg.Type_Promise:
		if p.data.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
			p.nodesResponded++
			//handles case where acceptor accepted previous value
			if rec.GetPreviousId() != 0 && rec.GetPreviousId() > p.data.GetPreviousId() { //checks if previously accepted value is newest one seen
				p.data.Value = rec.GetValue()
				p.data.PreviousId = rec.GetPreviousId()
			}
		}
	case msg.Type_Accept:
		if p.data.GetType() == msg.Type_Propose {
			p.nodesResponded++
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
