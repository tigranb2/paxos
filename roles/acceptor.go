package roles

import (
	"context"
	"errors"
	"paxos/msg"
	"time"
)

type Acceptor struct {
	acceptorData map[int32]*consensusData
	connections  map[int]msg.MessengerClient //connections to proposers
	ip           string                      //socket for Acceptor server
	ips          []string                    //sockets of all Proposers
}

type consensusData struct {
	promised int64
	accepted string
}

func InitAcceptor(ip string, ips []string) Acceptor {
	return Acceptor{acceptorData: make(map[int32]*consensusData), connections: createConnections(ips), ip: ip, ips: ips}
}

func (a *Acceptor) Run() {
	serverInit(a)
}

func (a *Acceptor) acceptMsg(rec *msg.Msg) (r *msg.Msg, err error) {
	if _, ok := a.acceptorData[rec.GetSlotIndex()]; !ok {
		a.acceptorData[rec.GetSlotIndex()] = &consensusData{}
	}

	acceptorSlotData := a.acceptorData[rec.GetSlotIndex()]
	switch rec.GetType() {
	case msg.Type_Prepare:
		//promise to ignore ids lower than received id
		if rec.GetId() > acceptorSlotData.promised {
			if acceptorSlotData.accepted != "" { //case where node has accepted prior value
				acceptorSlotData.promised = rec.GetId()
				r = &msg.Msg{Type: msg.Type_Promise, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId(), Value: acceptorSlotData.accepted, PreviousId: acceptorSlotData.promised}
			} else {
				acceptorSlotData.promised = rec.GetId()
				r = &msg.Msg{Type: msg.Type_Promise, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId()}
			}
		} else {
			err = errors.New("promised higher id")
		}
	case msg.Type_Propose:
		//accept proposed value
		if acceptorSlotData.promised == rec.GetId() {
			acceptorSlotData.accepted = rec.GetValue()
			r = &msg.Msg{Type: msg.Type_Accept, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId(), Value: rec.GetValue()}
			a.broadcast(&msg.SlotValue{Type: msg.Type_Accept, SlotIndex: rec.GetSlotIndex(), Value: rec.GetValue()}, int(rec.GetProposerId()))
		} else {
			err = errors.New("promised different id")
		}
	}
	return r, err
}

func (a *Acceptor) broadcast(m *msg.SlotValue, recProposer int) {
	for proposerId := range a.ips {
		if recProposer != proposerId+1 { //does not broadcast to proposer whose value was accepted
			//dials server if connection does not already exist
			if _, ok := a.connections[proposerId+1]; !ok {
				if c := dialServer(a.ips[proposerId]); c != nil {
					a.connections[proposerId+1] = c //save connection to acceptor's map
				} else {
					continue
				}
			}

			a.callProposer(proposerId+1, m)
		}
	}
}

func (a *Acceptor) callProposer(proposerId int, m *msg.SlotValue) {
	c := a.connections[proposerId]
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	//Send message to learner with 1 sec timeout
	_, err := c.MsgLearner(ctx, m)
	if err != nil {
		return
	}
}
