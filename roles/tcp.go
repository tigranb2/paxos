package roles

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

			//write data
			if err = bufWrite(t.writer, data); err != nil {
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
		n, err := bufRead(t.reader, readBuf)
		if err != nil {
			log.Fatalln("error reading from reader: ", err)
		}

		switch t.connType {
		case msg.ToAcceptor:
			var rec = &msg.Msg{}
			if err = rec.Unmarshal(readBuf[:n]); err == nil {
				t.receiveChan <- rec
			} else {
				log.Fatalln("error unmarshalling message: ", err)
			}
		case msg.ToProposer:
			var rec = &msg.SlotValue{}
			if err = rec.Unmarshal(readBuf[:n]); err == nil {
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

		go func() {
			reader, writer := initReaderWriter(conn)
			readBuf := make([]byte, 4096*100)

			writeChan := make(chan *msg.SlotValue)
			p.clientWriteChans[clientId] = writeChan

			for {
				var rec = &msg.QueueRequest{}
				//read QueueRequest from client
				n, errF := bufRead(reader, readBuf)
				if errF != nil {
					log.Fatalln("error reading from reader: ", errF)
				}

				if errF = rec.Unmarshal(readBuf[:n]); errF != nil {
					log.Fatalln("error unmarshalling value: ", errF)
				}
				rec.FromClient = clientId //save client id to request
				p.clientRequest <- rec

				//send SlotValue to client
				resp := <-writeChan

				data, errF := resp.Marshal()
				if errF = bufWrite(writer, data); errF != nil {
					log.Fatalln("error writing data: ", errF)
				}

				if errF = writer.Flush(); errF != nil {
					log.Fatalln("error flushing writer: ", errF)
				}
			}
		}()
	}
}

//acceptorServer receives messages from proposers
func (a *Acceptor) acceptorServer() {
	listener := initListener(a.ip)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("error accepting connection: ", err)
		}

		go func() {
			reader, writer := initReaderWriter(conn)
			readBuf := make([]byte, 4096*100)

			/*
				//reads proposerId, used to identify channel for accessing proposer writer
				//not needed for current design

				writeChan := make(chan *msg.Msg)
				var idMsg = &msg.Msg{} //msg from Proposer that will contain its ID

				n, errF := bufRead(reader, readBuf)
				if errF != nil {
					log.Fatalln("error reading from reader: ", errF)
				}
				if errF = idMsg.Unmarshal(readBuf[:n]); errF != nil {
					log.Fatalln("error unmarshalling value: ", errF)
				}
				a.proposerWriteChans[idMsg.GetProposerId()] = writeChan
			*/

			for {
				var rec = &msg.Msg{}
				//read msg from proposer
				n, errF := bufRead(reader, readBuf)
				if errF != nil {
					log.Fatalln("error reading from reader: ", errF)
				}

				if errF = rec.Unmarshal(readBuf[:n]); errF != nil {
					log.Fatalln("error unmarshalling value: ", errF)
				}

				resp := a.acceptMsg(rec)

				//write response to proposer
				data, errF := resp.Marshal()
				if errF = bufWrite(writer, data); errF != nil {
					log.Fatalln("error writing data: ", errF)
				}

				if errF = writer.Flush(); errF != nil {
					log.Fatalln("error flushing writer: ", errF)
				}
			}
		}()
	}
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

//bufWrite from https://github.com/haochenpan/rabia
func bufWrite(writer *bufio.Writer, data []byte) error {
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(data)))
	n1, err := writer.Write(lenBuf)
	if n1 != 4 {
		panic(fmt.Sprint("should not happen", err))
	}
	if err != nil {
		return err
	}
	n2, err := writer.Write(data)
	if n2 != len(data) {
		panic(fmt.Sprint("should not happen", err))
	}
	return err
}

//bufRead from https://github.com/haochenpan/rabia
func bufRead(reader *bufio.Reader, data []byte) (int, error) {
	lenBuf := make([]byte, 4)
	n1, err := io.ReadFull(reader, lenBuf)
	if n1 != 4 || err != nil {
		return n1, err
	}
	n2 := binary.LittleEndian.Uint32(lenBuf)
	n3, err := io.ReadFull(reader, data[:n2])
	if int(n2) != n3 {
		panic(fmt.Sprint("should not happen", err, int(n2), n3))
	}
	return n3, nil
}
