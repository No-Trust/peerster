// Routing Table implementation
package main

import (
	"github.com/No-Trust/peerster/common"
	"net"
	"sync"
	"time"
)

type RoutingTable struct {
	table       map[string]string // id -> ip:port
	mutex       *sync.Mutex
	peerID      string
	peerAddress string
}

func (r *RoutingTable) Get(id string) string {
	r.mutex.Lock()
	out := r.table[id]
	r.mutex.Unlock()
	return out
}

func NewRoutingTable(peerID string, peerAddr string) *RoutingTable {
	r := &RoutingTable{
		table:       make(map[string]string),
		mutex:       &sync.Mutex{},
		peerID:      peerID,
		peerAddress: peerAddr,
	}
	return r
}

func (r *RoutingTable) Str() string {
	str := ""
	r.mutex.Lock()
	for id, ipport := range r.table {
		str = str + "\t" + id + " " + ipport + "\n"
	}
	r.mutex.Unlock()
	return str
}

func (r *RoutingTable) AddNextHop(origin string, remoteaddr *net.UDPAddr) {
	remoteaddrStr := UDPAddrToString(remoteaddr)

	if origin == r.peerID {
		return
	}

	if remoteaddrStr == r.peerAddress {
		return
	}
	
	if origin == "" {
		return
	}


	r.mutex.Lock()
	r.table[origin] = remoteaddrStr
	r.mutex.Unlock()
}

func (r *RoutingTable) copy() *RoutingTable {
	newR := NewRoutingTable(r.peerID, r.peerAddress)
	r.mutex.Lock()
	for key, value := range r.table {
		newR.table[key] = value
	}
	r.mutex.Unlock()

	return newR
}

func (r *RoutingTable) GetIds() []string {
	r.mutex.Lock()
	ids := make([]string, len(r.table))
	for k, _ := range r.table {
		ids = append(ids, k)
	}
	r.mutex.Unlock()
	return ids
}

// Implementation of the simplified DSDV routing algorithm.
func routerumor(g *Gossiper, rtimer uint, wg sync.WaitGroup) {
	/*
	 * Thread responsible for sending route rumor messages
	 * Sends a route rumor every rtimer seconds
	 */

	defer wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(rtimer)) // every rtimer sec
	defer ticker.Stop()

	for _ = range ticker.C {
		A := g.peerSet.RandomPeer()
		if A != nil {

			nextSeq := g.vectorClock.Get(g.Parameters.Identifier)

			routerumor := RumorMessage{
				Origin: g.Parameters.Identifier,
				ID:     nextSeq,
				Text:   "",
			}

			// update status vector
			g.vectorClock.Update(g.Parameters.Identifier)

			// update messages
			g.messages.Add(&routerumor)

			// send route rumor packet
			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					Rumor: &routerumor,
				},
				Destination: A.Address,
			}
		}
	}
}

// Brodcast a route rumor to some peers.
// TODO : this can be generic for any message
func (rumor *RumorMessage) broadcastTo(g *Gossiper, ps common.PeerSet) {
	// broadcast the route message to all given peers
	peers := ps.ToPeerArray()
	for _, peer := range peers {
		g.standardOutputQueue <- rumor.MongeringString(&peer.Address)
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				Rumor: rumor,
			},
			Destination: peer.Address,
		}
	}
}

// Create a route rumor and broadcast it
func broadcastNewRoute(g *Gossiper) {
	// broadcast a new route message to all known peers
	peers := g.peerSet.ToPeerArray()
	for _, peer := range peers {

		nextSeq := g.vectorClock.Get(g.Parameters.Identifier)

		routerumor := RumorMessage{
			Origin: g.Parameters.Identifier,
			ID:     nextSeq,
			Text:   "",
		}

		g.standardOutputQueue <- routerumor.MongeringString(&peer.Address)

		// update status vector
		g.vectorClock.Update(g.Parameters.Identifier)

		// update messages
		g.messages.Add(&routerumor)

		// send route rumor packet
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				Rumor: &routerumor,
			},
			Destination: peer.Address,
		}
	}
}
