package main

import (
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
)

type ENRData struct {
	ID      string
	IP      net.IP
	UDPPort int
	TCPPort int
}

// 1. On crée une entrée générique
type GenericEntry struct {
	KeyName string
	Value   rlp.RawValue // Garde les bytes RLP bruts sans les transformer
}

func (g *GenericEntry) ENRKey() string { return g.KeyName }

// ExtractENRData pulls specific identity and networking fields from a record
func ExtractENRData(record *enr.Record) ENRData {
	data := ENRData{}

	// 2. Extract IPv4
	var ip4 enr.IPv4
	if err := record.Load(&ip4); err == nil {
		data.IP = net.IP(ip4)
	}

	// 3. Extract UDP Port (Usually for discovery)
	var udp enr.UDP
	if err := record.Load(&udp); err == nil {
		data.UDPPort = int(udp)
	}

	// 4. Extract TCP Port (Usually for the actual devp2p connection)
	var tcp enr.TCP
	if err := record.Load(&tcp); err == nil {
		data.TCPPort = int(tcp)
	}
	var id enr.ID
	if err := record.Load(&id); err == nil {
		data.ID = string(id)
	}
	elements := record.AppendElements(nil)
	for i, v := range elements {
		if i%2 == 1 {
			switch v {
			case "eth":
				b := elements[i+1].(rlp.RawValue)
				_ = b
			}
			fmt.Println(v)
		}
	}
	return data
}
