package main

import (
	"fmt"
	"log"
	"os"
	"paxos/config"
	"paxos/network"
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
	quorum := configData.GetQuorum()
	acceptorSockets := configData.ParseSockets("acceptor")
	proposerSockets := configData.ParseSockets("proposer")

	switch arguments[1] {
	case "a":
		a := network.InitAcceptor(acceptorSockets[nodeId-1])
		a.Run()
	case "p":
		p := network.InitProposer(acceptorSockets, nodeId, proposerSockets[nodeId-1], quorum)
		p.Run()
	case "c":
		c := network.InitClient(proposerSockets[nodeId-1], configData.Timeout)
		c.CloseLoopClient()
	}

}
