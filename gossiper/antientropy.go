// Implementation of the anti entropy algorithm.
package main

import (
	"time"

	"github.com/No-Trust/peerster/common"
)

// Implementation of the anti entropy algorithm.
// Send a status packet every etimer seconds, to a random peer.
func antiEntropy(g *Gossiper, etimer uint) {

	ticker := time.NewTicker(time.Second * time.Duration(etimer)) // every rate sec
	defer ticker.Stop()

	for range ticker.C {

		randPeer := g.reputationTable.ContribRandomPeer()

		if randPeer != "" {
			// send status packet

			addr := stringToUDPAddr(randPeer)

			status := g.vectorClock.Copy()
			common.Log(status.AntiEntropyString(&addr), common.LOG_MODE_FULL)

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					Status: status,
				},
				Destination: addr,
			}
		}
	}
}
