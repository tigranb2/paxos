package network

import (
	"errors"
	"paxos/msg"
)

type Acceptor struct {
	acceptorData map[int32]*consensusData
	ip           string //socket for Acceptor server
}

type consensusData struct {
	promised int64
	accepted string
}

func InitAcceptor(ip string) Acceptor {
	return Acceptor{ip: ip, acceptorData: make(map[int32]*consensusData)}
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
		} else {
			err = errors.New("promised different id")
		}
	}
	return r, err
}
