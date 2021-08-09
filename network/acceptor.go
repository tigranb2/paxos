package network

import (
	"errors"
	"paxos/msg"
)

type Acceptor struct {
	Ip       string
	promised int64
	accepted string
}

func InitAcceptor(ip string) Acceptor {
	return Acceptor{Ip: ip}
}

func (a *Acceptor) Run() {
	serverInit(a)
}

func (a *Acceptor) acceptMsg(rec *msg.Msg) (*msg.Msg, error) {
	switch rec.GetType() {
	case msg.Type_Prepare:
		//promise to ignore ids lower than received id
		if rec.GetId() > a.promised {
			if a.accepted != "" { //case where node has accepted prior value
				resp := msg.Msg{Type: msg.Type_Promise, Id: rec.GetId(), Value: a.accepted, PreviousId: a.promised}
				a.promised = rec.GetId()
				return &resp, nil
			} else {
				resp := msg.Msg{Type: msg.Type_Promise, Id: rec.GetId()}
				a.promised = rec.GetId()
				return &resp, nil
			}
		}
	case msg.Type_Propose:
		//accept proposed value
		if a.promised == rec.GetId() {
			a.accepted = rec.GetValue()
			resp := msg.Msg{Type: msg.Type_Accept, Id: rec.GetId(), Value: rec.GetValue()}
			return &resp, nil
		}
	}

	return nil, errors.New("placeholder") //simulate timeout by returning error
}
