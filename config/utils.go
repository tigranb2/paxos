package config
// useful functions for config data

import (
	"log"
	"strconv"
	"strings"
)

func (configData *config) GetQuorum() int {
	return configData.Acceptors.Count/2 + 1
}

// ParseNetworkRanges parses ip and port ranges and checks for valid input
func (configData *config) ParseNetworkRanges() (CONNECTIONS []string) {
	var err error
	ipMin, ipMax, portMin, portMax := 0, 0, 0, 0

	ipRange := strings.Split(configData.Acceptors.ipRange, "-")
	if len(ipRange) > 1 {
		ipMax, err = strconv.Atoi(ipRange[1])
	}
	ipMin, err = strconv.Atoi(ipRange[0])
	if err != nil {
		log.Fatalf("ipRange not formatted correctly: %v", err)
	}

	portRange := strings.Split(configData.Acceptors.portRange, "-")
	if len(portRange) > 1 {
		portMax, err = strconv.Atoi(portRange[1])
	}
	portMin, err = strconv.Atoi(portRange[0])
	if err != nil {
		log.Fatalf("portRange not formatted correctly: %v", err)
	}

	if ipMax-ipMin < configData.Acceptors.Count {
		log.Fatalf("ipRange too small")
	}
	if portMax-portMin < configData.Acceptors.Count {
		log.Fatalf("portRange too small")
	}

	for i := 0; i < configData.Acceptors.Count; i++ {
		ip := strconv.Itoa(ipMin + ipMax)
		port := strconv.Itoa(portMin + portMax)
		CONNECTIONS = append(CONNECTIONS, ip+":"+port)
	}

	return CONNECTIONS
}