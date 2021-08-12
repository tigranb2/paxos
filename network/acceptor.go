package network

import (
	"paxos/msg"
)

type Acceptor struct {
	Ip           string
	acceptorData map[int32]*consensusData
}

type consensusData struct {
	promised int64
	accepted string
}

func InitAcceptor(ip string, instances int) Acceptor {
	aData := make(map[int32]*consensusData)
	for i := 1; i <= instances; i++ {
		aData[int32(i)] = &consensusData{}
	}
	return Acceptor{Ip: ip, acceptorData: aData}
}

func (a *Acceptor) Run() {
	serverInit(a)
}

func (a *Acceptor) acceptMsg(rec *msg.MsgSlice) *msg.MsgSlice {
	resp := msg.MsgSlice{Msgs: make(map[int32]*msg.Msg)}
	for instance, recMsg := range rec.Msgs {
		aData := a.acceptorData[instance]
		var r msg.Msg
		switch recMsg.GetType() {
		case msg.Type_Prepare:
			//promise to ignore ids lower than received id
			if recMsg.GetId() > aData.promised {
				if aData.accepted != "" { //case where node has accepted prior value
					aData.promised = recMsg.GetId()
					r = msg.Msg{Type: msg.Type_Promise, Id: recMsg.GetId(), Value: aData.accepted, PreviousId: aData.promised}
				} else {
					aData.promised = recMsg.GetId()
					r = msg.Msg{Type: msg.Type_Promise, Id: recMsg.GetId()}
				}
			} else {
				r = msg.Msg{Type: msg.Type_Promise, Id: 0}
			}
		case msg.Type_Propose:
			//accept proposed value
			if aData.promised == recMsg.GetId() {
				aData.accepted = recMsg.GetValue()
				r = msg.Msg{Type: msg.Type_Accept, Id: recMsg.GetId(), Value: recMsg.GetValue()}
			} else {
				r = msg.Msg{Type: msg.Type_Promise, Id: 0}
			}
		}
		resp.Msgs[instance] = &r
	}
	return &resp
}
