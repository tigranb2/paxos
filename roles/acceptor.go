package roles

type Acceptor struct {
	ip string
}

func InitAcceptor(ip string) Acceptor {
	return Acceptor{ip: ip}
}

func (a *Acceptor) Run() {
	a.acceptorServer()
}
