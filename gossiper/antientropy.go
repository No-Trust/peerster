// Implementation of the anti entropy algorithm.
package main

import (
	"sync"
	"time"
)

// Implementation of the anti entropy algorithm.
// Send a status packet every etimer seconds, to a random peer.
func antiEntropy(g *Gossiper, etimer uint, wg sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(etimer)) // every rate sec
	defer ticker.Stop()

	for _ = range ticker.C {
		A := g.peerSet.RandomPeer()
		if A != nil {
			// send status packet
			status := &g.vectorClock
			g.standardOutputQueue <- status.AntiEntropyString(&A.Address)

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					Status: status,
				},
				Destination: A.Address,
			}
		}
	}
}
