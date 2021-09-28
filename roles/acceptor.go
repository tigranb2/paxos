package roles

import (
	"paxos/msg"
)

type Acceptor struct {
	acceptorData       map[int32]*consensusData
	int64Needed        int      //number of int64s needed
	ip                 string   //socket for Acceptor server
	ips                []string //sockets of all Proposers
	proposerWriteChans map[int32]chan *msg.Msg
}

type consensusData struct {
	promised int64
	accepted string
}

func InitAcceptor(int64Needed int, ip string, ips []string) Acceptor {
	return Acceptor{acceptorData: make(map[int32]*consensusData), int64Needed: int64Needed, ip: ip, ips: ips, proposerWriteChans: make(map[int32]chan *msg.Msg)}
}

func (a *Acceptor) Run() {
	a.acceptorServer()
}

func (a *Acceptor) acceptMsg(rec *msg.Msg) (r *msg.Msg) {
	if _, ok := a.acceptorData[rec.GetSlotIndex()]; !ok {
		a.acceptorData[rec.GetSlotIndex()] = &consensusData{}
	}

	acceptorSlotData := a.acceptorData[rec.GetSlotIndex()]
	switch rec.GetType() {
	case msg.Prepare:
		//promise to ignore ids lower than received id
		if rec.GetId() > acceptorSlotData.promised {
			if acceptorSlotData.accepted != "" { //case where node has accepted prior value
				acceptorSlotData.promised = rec.GetId()
				r = &msg.Msg{Type: msg.Promise, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId(), Value: acceptorSlotData.accepted, PreviousId: acceptorSlotData.promised, Size_: make([]int64, a.int64Needed)}
			} else {
				acceptorSlotData.promised = rec.GetId()
				r = &msg.Msg{Type: msg.Promise, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId(), Size_: make([]int64, a.int64Needed)}
			}
		} else {
			r = nil //promised higher id
		}
	case msg.Propose:
		//accept proposed value
		if acceptorSlotData.promised == rec.GetId() {
			acceptorSlotData.accepted = rec.GetValue()
			r = &msg.Msg{Type: msg.Accept, SlotIndex: rec.GetSlotIndex(), Id: rec.GetId(), Value: rec.GetValue(), Size_: make([]int64, a.int64Needed)}

			//send to all learners except for learner of proposer who sent the message
			for i := 1; i <= len(a.proposerWriteChans); i++ {
				if int32(i) != rec.GetProposerId() {
					a.proposerWriteChans[int32(i)] <- &msg.Msg{Type: msg.LearnerMsg, SlotIndex: rec.GetSlotIndex(), Value: rec.GetValue(), Size_: make([]int64, a.int64Needed)}
				}
			}
		} else {
			r = nil //promised different id
		}
	}
	return r
}
