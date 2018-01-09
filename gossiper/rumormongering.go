// Implementation of the rumormongering algorithm.
package main

import (
	"github.com/No-Trust/peerster/common"
	"time"
)

// Implementation of the rumormongering algorithm.
// Send a rumor to destination and continue with a random peer with probability 1/2
func (g *Gossiper) rumormonger(rumor *RumorMessage, destPeer *common.Peer) {
	g.standardOutputQueue <- rumor.MongeringString(destPeer.Address)
	// send rumor to peer

	g.gossipOutputQueue <- &Packet{
		GossipPacket: GossipPacket{
			Rumor: rumor,
		},
		Destination: destPeer.Address,
	}

  // Decrease contribution-based reputation of receiver
  g.reputationTable.DecreaseContribRep(addrToString(destPeer.Address))

	// and wait for status message
	statusChannel := make(chan *PeerStatus)
	// format : id/ip:port/nextID

	ackID := AckString(destPeer.Address, rumor.Origin, rumor.ID+1)

	// register for listener
	g.waitersMutex.Lock()
	_, present := g.gossiperWaiters[ackID]
	g.waitersMutex.Unlock()

	if present {
		// there is a goroutine already waiting for this status message

		// too bad
		return
	}

	// Register
	g.waitersMutex.Lock()
	g.gossiperWaiters[ackID] = statusChannel
	g.waitersMutex.Unlock()

	// waiter
	go func() {
		timer := time.NewTimer(time.Millisecond * 1000)

		// stop when the first channel is ready : either timeout or receive ack
		select {
		case <-timer.C:
			// timer stops first
			// timeout
			timer.Stop()
			g.waitersMutex.Lock()
			close(statusChannel)
			g.gossiperWaiters[ackID] = nil
			delete(g.gossiperWaiters, ackID)
			g.waitersMutex.Unlock()

      randPeer := g.reputationTable.ContribRandomPeer()

			if randPeer != "" {
				go g.rumormonger(rumor, &common.Peer {
				  Address : stringToUDPAddr(randPeer),
				})
			}

			return

		case <-statusChannel:
			// received the ack before timeout
			timer.Stop()
			g.waitersMutex.Lock()
			close(statusChannel)
			g.gossiperWaiters[ackID] = nil
			delete(g.gossiperWaiters, ackID)
			g.gossiperWaiters[ackID] = nil
			g.waitersMutex.Unlock()
			// rumormonger again with probability 1/2
			if flipCoin() {
				g.standardOutputQueue <- CoinFlipString(&destPeer.Address)

        randPeer := g.reputationTable.ContribRandomPeer()

				if randPeer != "" {
					go g.rumormonger(rumor, &common.Peer {
					  Address : stringToUDPAddr(randPeer),
					})
				}

			}
			return
		}
	}()

	return
}
