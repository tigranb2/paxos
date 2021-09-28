package main

import (
	"fmt"
	"log"
	"os"
	"paxos/config"
	"paxos/roles"
	"strconv"
)

func main() {
	arguments := os.Args
	if len(arguments) < 3 {
		fmt.Println("Please specify node type (p, a) and node id")
		return
	}
	nodeId, err := strconv.Atoi(arguments[2])
	if err != nil {
		log.Fatalf("node id formatted incorrectly: %v", err)
	}

	configData := config.ParseConfig()
	acceptorSockets := configData.ParseSockets("acceptor")
	proposerSockets := configData.ParseSockets("proposer")

	bytesNeeded := (configData.MsgSize*1024)/8 - 6
	if bytesNeeded <= 0 {
		bytesNeeded = 1
	}

	switch arguments[1] {
	case "a":
		a := roles.InitAcceptor(bytesNeeded, acceptorSockets[nodeId-1], proposerSockets)
		a.Run()
	case "p":
		p := roles.InitProposer(bytesNeeded, acceptorSockets, nodeId, proposerSockets[nodeId-1], configData.Quorum)
		p.Run()
	case "c":
		c := roles.InitClient(nodeId, proposerSockets)
		c.CloseLoopClient()
	}

}

// ! Implement !
func determineInt32Needed(msgSize int) int {
	msgSize -= 52 //other fields of msg take up 52 bytes of space

	if r := msgSize % 4; r != 0 { //change msgSize into a multiple of 4 (floor function)
		msgSize -= r
	}

	if msgSize <= 13 {
		return 0
	}
	return msgSize / 4
}
