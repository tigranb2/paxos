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
	connections := configData.ParseNetworkRanges()

	switch arguments[1] {
	case "p":
		p := network.InitProposer(configData.Proposers.Values[nodeId-1], connections)
		p.Run(quorum)
	case "a":
		a := network.InitAcceptor(connections[nodeId-1])
		a.Run()
	}
}
