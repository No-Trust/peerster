// PeerSet implementation : a set of peers
package common

import (
	"math/rand"
	"net"
	"strconv"
	"sync"
)

type PeerSlice struct {
	Peers []Peer
}

// PeerSet, a set of peers
// Thread Safe
type PeerSet struct {
	peers []Peer
	mutex *sync.Mutex
	except net.UDPAddr // address to exclude : own's address
}

// Add a Peer to a PeerSet
func (ps *PeerSet) Add(peer Peer) {
	// check if exists
	if !ps.Contains(peer) {
		// if it does not exists yet, add it
		ps.mutex.Lock()
		// if this is an ipv6 address, do not add
		if peer.Address.IP.To4() != nil {
			// this is an IPv4 address
			if ps.except.IP.String() != peer.Address.IP.String() || ps.except.Port != peer.Address.Port {
				ps.peers = append(ps.peers, peer.Copy())
			}
		}
		ps.mutex.Unlock()
	}
}

// Check if a PeerSet contains a Peer
func (ps PeerSet) Contains(peer Peer) bool {
	ps.mutex.Lock()
	for _, p := range ps.peers {
		if (&p).Equals(&peer) {
			ps.mutex.Unlock()
			return true
		}
	}
	ps.mutex.Unlock()
	return false
}

func (ps PeerSet) Remove(peer *Peer) PeerSet {
	newPeers := ps.ToPeerArray()
	// look for the peer
	index := -1
	for i, p := range newPeers {
		if p.Equals(peer) {
			index = i
			break
		}
	}

	if index >= 0 {
		// remove peer
		l := len(newPeers)
		// swaps the last element and the element to be removed
		newPeers[l-1], newPeers[index] = newPeers[index], newPeers[l-1]
		newPeers = newPeers[:l-1]

		return PeerSet{
			peers: newPeers,
			mutex: &sync.Mutex{},
		}
	}
	return PeerSet{
		peers: ps.ToPeerArray(),
		mutex: &sync.Mutex{},
	}
}

func (ps PeerSet) ToPeerArray() []Peer {
	ps.mutex.Lock()
	slice := make([]Peer, len(ps.peers))
	for i, peer := range ps.peers {
		slice[i] = peer.Copy()
	}
	ps.mutex.Unlock()
	return slice
}

func (ps PeerSet) ToPeerSlice() PeerSlice {
	slice := ps.ToPeerArray()
	return PeerSlice{slice}
}

func NewSet(except net.UDPAddr) PeerSet {
	return PeerSet{
		peers: make([]Peer, 0),
		mutex: &sync.Mutex{},
		except: except,
	}
}

func NewSetFromAddrs(addrs []net.UDPAddr, except net.UDPAddr) PeerSet {
	newPeerSet := NewSet(except)
	for _, addr := range addrs {
		var p Peer = Peer{addr, ""}
		newPeerSet.Add(p)
	}
	return newPeerSet
}

func (ps PeerSet) RandomPeer() *Peer {
	// Returns a random peer from the PeerSet
	ps.mutex.Lock()
	if len(ps.peers) == 0 {
		ps.mutex.Unlock()
		return nil
	}

	i := rand.Int() % len(ps.peers)
	rpeer := ps.peers[i]
	npeer := rpeer.Copy()
	ps.mutex.Unlock()
	return &npeer
}

func (ps PeerSet) PeersListString() *string {
	str := ""
	ps.mutex.Lock()
	for _, peer := range ps.peers {
		if str != "" {
			str = str + ","
		}
		str = str + peer.Address.IP.String() + ":" + strconv.Itoa(peer.Address.Port)
	}
	ps.mutex.Unlock()
	return &str
}

func (ps PeerSet) Str() string {
	str := "PeerSet :\n"
	ps.mutex.Lock()
	for _, peer := range ps.peers {
		str = str + "\t" + peer.Str() + "\n"
	}
	ps.mutex.Unlock()
	return str
}
