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
	quorum := configData.GetQuorum()

	acceptorSockets := configData.ParseSockets("acceptor")
	proposerSockets := configData.ParseSockets("proposer")

	switch arguments[1] {
	case "a":
		a := roles.InitAcceptor(acceptorSockets[nodeId-1], proposerSockets)
		a.Run()
	case "p":
		p := roles.InitProposer(acceptorSockets, nodeId, proposerSockets[nodeId-1], quorum)
		p.Run()
	case "c":
		c := roles.InitClient(nodeId, proposerSockets, configData.Timeout)
		c.CloseLoopClient()
	}

}
