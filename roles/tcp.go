package roles

import (
	"bufio"
	"io"
	"log"
	"net"
	"paxos/msg"
	"time"
)

type TCP struct {
	conn        *net.Conn
	connType    msg.ConnType
	ip          string        //ip that will be connected to
	reader      *bufio.Reader //reader bound to conn
	receiveChan chan interface{}
	sendChan    chan interface{}
	writer      *bufio.Writer //writer bound to conn

}

func initTCP(connType msg.ConnType, ip string, receiveChan chan interface{}) *TCP {
	t := TCP{connType: connType, ip: ip, receiveChan: receiveChan, sendChan: make(chan interface{})}

	err := t.connect()
	for err != nil {
		time.Sleep(time.Second)
		err = t.connect()
	}

	return &t
}

func (t *TCP) connect() error {
	if conn, err := net.Dial("tcp", t.ip); err == nil {
		t.conn = &conn
		t.reader, t.writer = initReaderWriter(conn)
		return err
	} else {
		return err
	}
}

func (t *TCP) sendMsgs() {
	for {
		select {
		case req := <-t.sendChan:
			var data []byte
			var err error

			switch t.connType {
			case msg.ToAcceptor:
				if m, ok := req.(*msg.Msg); ok {
					data, err = m.Marshal()
				} else {
					log.Fatalln("error converting message to proper type")
				}
			case msg.ToProposer:
				if m, ok := req.(*msg.QueueRequest); ok {
					data, err = m.Marshal()
				} else {
					log.Fatalln("error converting message to proper type")
				}
			case msg.ToLearner:
				if m, ok := req.(*msg.SlotValue); ok {
					data, err = m.Marshal()
				} else {
					log.Fatalln("error converting message to proper type")
				}
			}

			if err != nil {
				log.Fatalln("error occurred marshalling request: ", err)
			}

			if _, err = t.writer.Write(data); err != nil {
				log.Fatalln("error writing data: ", err)
			}

			if err = t.writer.Flush(); err != nil {
				log.Fatalln("error flushing writer: ", err)
			}
		}
	}
}

func (t *TCP) receiveMsgs() {
	readBuf := make([]byte, 4096*100)
	for {
		_, err := io.ReadFull(t.reader, readBuf)
		if err != nil {
			log.Fatalln("error reading from reader: ", err)
		}

		switch t.connType {
		case msg.ToAcceptor:
			var rec *msg.Msg
			if err := rec.Unmarshal(readBuf); err == nil {
				t.receiveChan <- rec
			} else {
				log.Fatalln("error unmarshalling message: ", err)
			}
		case msg.ToProposer:
			var rec *msg.SlotValue
			if err := rec.Unmarshal(readBuf); err == nil {
				t.receiveChan <- rec
			} else {
				log.Fatalln("error unmarshalling message: ", err)
			}
		}
	}
}

//proposerServer receives messages from clients
func (p *Proposer) proposerServer() {
	listener := initListener(p.ip)
	var clientId int32
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("error accepting connection: ", err)
		}

		clientId++
		p.clientWriteChans[clientId] = make(chan *msg.SlotValue)

		go func() {
			reader, writer := initReaderWriter(conn)
			readBuf := make([]byte, 4096*100)

			for {
				var rec *msg.QueueRequest
				//read QueueRequest from client
				_, errF := io.ReadFull(reader, readBuf)
				if errF != nil {
					log.Fatalln("error reading from reader: ", errF)
				}

				if errF = rec.Unmarshal(readBuf); errF == nil {
					p.clientRequest <- rec
				}

				//send SlotValue to client
				resp := <-p.clientWriteChans[clientId]
				data, errF := resp.Marshal()

				if _, err = writer.Write(data); err != nil {
					log.Fatalln("error writing data: ", err)
				}

				if err = writer.Flush(); err != nil {
					log.Fatalln("error flushing writer: ", err)
				}
			}
		}()
	}
}

//acceptorServer receives messages from proposers
func (a *Acceptor) acceptorServer() {

}

//learnerServer receives messages from acceptors
func (p *Proposer) learnerServer() {
}

func initListener(ip string) net.Listener {
	listener, err := net.Listen("tcp", ip)
	if err != nil {
		log.Println("error opening listener: ", err)
	}

	return listener
}

func initReaderWriter(conn net.Conn) (*bufio.Reader, *bufio.Writer) {
	err := (conn).(*net.TCPConn).SetWriteBuffer(4096000)
	if err != nil {
		log.Fatalln(err)
	}
	err = (conn).(*net.TCPConn).SetReadBuffer(4096000)
	if err != nil {
		log.Fatalln(err)
	}
	err = (conn).(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		log.Fatalln(err)
	}
	err = (conn).(*net.TCPConn).SetKeepAlivePeriod(20 * time.Second)
	if err != nil {
		log.Fatalln(err)
	}

	return bufio.NewReaderSize(conn, 4096000), bufio.NewWriterSize(conn, 4096000)
}
