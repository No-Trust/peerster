// Peer implementation
package common

import (
	"fmt"
	"net"
)

// A gossiper Peer
type Peer struct {
	Address    net.UDPAddr
	Identifier string
}

// (Deep) Copy of a Peer
func (peer *Peer) Copy() Peer {
	naddr := net.UDPAddr{
		IP:   net.ParseIP(peer.Address.IP.String()),
		Port: peer.Address.Port,
		Zone: "",
	}
	npeer := Peer{
		Address:    naddr,
		Identifier: peer.Identifier,
	}

	return npeer
}

func (peer *Peer) Str() string {
	str := fmt.Sprintf("%s@%s:%d%s", peer.Identifier, peer.Address.IP.String(), peer.Address.Port, peer.Address.Zone)
	return str
}

func (A *Peer) Equals(B *Peer) bool {
	return A.Str() == B.Str()
}
