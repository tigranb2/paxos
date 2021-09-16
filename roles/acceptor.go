package roles

import (
	"paxos/msg"
)

type Acceptor struct {
	acceptorData map[int32]*consensusData
	int64Needed  int      //number of int64s needed
	ip           string   //socket for Acceptor server
	ips          []string //sockets of all Proposers
}

type consensusData struct {
	promised int64
	accepted string
}

func InitAcceptor(int64Needed int, ip string, ips []string) Acceptor {
	return Acceptor{acceptorData: make(map[int32]*consensusData), int64Needed: int64Needed, ip: ip, ips: ips}
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

			//send to Learners!
			//a.broadcast(&msg.SlotValue{Type: msg.Accept, SlotIndex: rec.GetSlotIndex(), Value: rec.GetValue()}, int(rec.GetProposerId()))
		} else {
			r = nil //promised different id
		}
	}
	return r
}

/*
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//Send message to learner with 1 sec timeout
	_, err := c.MsgLearner(ctx, m)
	if err != nil {
		return
	}
}
*/
