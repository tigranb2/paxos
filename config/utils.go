package config

//useful functions for config data

import (
	"log"
)

//GetQuorum returns number of Acceptors needed for majority
func (configData *Config) GetQuorum() int {
	return configData.Acceptors.Count/2 + 1
}

func (configData *Config) ParseSockets(nodeType string) (connections []string) {
	node := configData.Acceptors
	if nodeType == "proposer" {
		node = configData.Proposers
	}

	if len(node.Ip) != len(node.Port) {
		log.Fatalf("number of ips does not equal number of ports")
	}

	for i := 0; i < len(node.Ip); i++ {
		connections = append(connections, node.Ip[i]+":"+node.Port[i])
	}

	return connections
}

/*
//For testing
//ParseSockets parses socket info from config for either Acceptors or Proposers
func (configData *Config) ParseSocketsTesting(nodeType string) (connections []string) {
	node := configData.Acceptors
	if nodeType == "proposer" {
		node = configData.Proposers
	}
	ip := net.ParseIP(node.InitIP)
	if ip == nil {
		log.Fatalf("initIP is not a valid ipv4 address")
	}
	port, err := strconv.Atoi(node.InitPort)
	if err != nil {
		log.Fatalf("initPort is not a valid port")
	}
	connections = append(connections, ip.String()+":"+strconv.Itoa(port))
	for i := 1; i < node.Count; i++ {
		if node.IncrementIP {
			ip = nextIP(ip, 1)
		}
		if node.IncrementPort {
			port++
		}
		connections = append(connections, ip.String()+":"+strconv.Itoa(port))
	}
	return connections
}
//nextIP returns the adjacent ipv4 address
func nextIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3)
}
*/
