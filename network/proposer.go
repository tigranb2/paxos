package network

import (
	"fmt"
	"math/rand"
	"paxos/msg"
	"time"
)

type Proposer struct {
	connections  map[string]msg.MessengerClient //store server connections for optimization
	converged    []string                       //stores values of converged instances
	data         *msg.MsgSlice                  //stores proposed id, value, and current phase for all instances
	quorum       int
	quorumsData  map[int32]*quorumData //keeps track of responses for quorum
	quorumStatus chan bool
}

type quorumData struct {
	promises, accepts int
	proposeAttempt    int //stores number of times proposer has attempted propose phase
}

func InitProposer(instanceCount int, value string, connections []string, quorum int) Proposer {
	data := msg.MsgSlice{}
	data.Msgs = make(map[int32]*msg.Msg)
	quorumsData := make(map[int32]*quorumData)
	for i := 1; i <= instanceCount; i++ { //initialize maps
		data.Msgs[int32(i)] = &msg.Msg{Type: msg.Type_Prepare, Id: time.Now().UnixNano(), Value: value}
		quorumsData[int32(i)] = &quorumData{}
	}
	return Proposer{connections: initConnectionsMap(connections), data: &data, quorum: quorum, quorumsData: quorumsData, quorumStatus: make(chan bool)}
}

func (p *Proposer) Run(stop chan []string) {
	for {
		select {
		case <-p.quorumStatus:
			for instance, m := range p.data.GetMsgs() {
				pData := p.quorumsData[instance]
				pMsg := p.data.Msgs[instance]
				if m.GetType() == msg.Type_Prepare {
					if pData.promises >= p.quorum { //prepare phase passed
						pMsg.Type = msg.Type_Propose
					} else { //prepare phase failed
						pData.promises = 0
						pMsg.Id = time.Now().UnixNano()
					}
				} else {
					if pData.accepts >= p.quorum { //propose phase passed
						p.converged = append(p.converged, fmt.Sprintf("Instance %v converged on %v", instance, pMsg.GetValue()))
						delete(p.data.Msgs, instance)
					} else if pData.proposeAttempt == 3 { //propose phase failed 3 times
						pData.proposeAttempt = 0
						pData.accepts = 0
						pMsg.Type = msg.Type_Prepare
						pMsg.Id = time.Now().UnixNano()
					} else { //propose phase failed less than 3 times
						pData.accepts = 0
						pData.proposeAttempt++
					}
				}
			}

			//terminate when all instances of the proposer converge
			if len(p.data.GetMsgs()) == 0 {
				stop <- p.converged
				return
			}

			go p.Broadcast()
		}
	}
}

func (p *Proposer) Broadcast() {
	for netIp := range p.connections {
		msgServer(netIp, p)
	}

	n := rand.Intn(1000)
	time.Sleep(time.Millisecond * time.Duration(n)) //sleep to break proposer ties
	p.quorumStatus <- true
}

func (p *Proposer) handleResponse(rec *msg.MsgSlice) {
	for instance, recMsg := range rec.Msgs {
		if recMsg.GetId() == 0 { //simulate timeout
			return
		}
		pMsg := p.data.Msgs[instance]
		pData := p.quorumsData[instance]
		switch recMsg.GetType() {
		case msg.Type_Promise:
			if pMsg.GetType() == msg.Type_Prepare { //check that proposer is asking for promises
				pData.promises++
				//handles case where acceptor accepted previous value
				if recMsg.GetPreviousId() != 0 && recMsg.GetPreviousId() > pMsg.GetPreviousId() { //checks if previously accepted value is newest one seen
					pMsg.Value = recMsg.GetValue()
					pMsg.PreviousId = recMsg.GetPreviousId()
				}
			}
		case msg.Type_Accept:
			if pMsg.GetType() == msg.Type_Propose {
				pData.accepts++
			}
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
