// Implementation of the anti entropy algorithm.
package main

import (
	"time"
)

// Implementation of the anti entropy algorithm.
// Send a status packet every etimer seconds, to a random peer.
func antiEntropy(g *Gossiper, etimer uint) {

	ticker := time.NewTicker(time.Second * time.Duration(etimer)) // every rate sec
	defer ticker.Stop()

	for _ = range ticker.C {

		randPeer := g.reputationTable.ContribRandomPeer()

		if randPeer != "" {
			// send status packet

			addr := stringToUDPAddr(randPeer)

			status := g.vectorClock.Copy()
			g.standardOutputQueue <- status.AntiEntropyString(&addr)

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					Status: status,
				},
				Destination: addr,
			}
		}
	}
}
