// Implementation of the rumormongering algorithm.
package main

import (
	"fmt"
	"github.com/No-Trust/peerster/common"
	"time"
)

// Implementation of the rumormongering algorithm.
// Send a rumor to destination and continue with a random peer with probability 1/2
func (g *Gossiper) rumormonger(rumor *RumorMessage, destPeer *common.Peer) {

	g.standardOutputQueue <- rumor.MongeringString(&destPeer.Address)
	// send rumor to peer

	g.gossipOutputQueue <- &Packet{
		GossipPacket: GossipPacket{
			Rumor: rumor,
		},
		Destination: destPeer.Address,
	}

	// and wait for status message
	statusChannel := make(chan *PeerStatus)
	// format : id/ip:port/nextID

	//statusString := rumor.Origin + "/" + addrToString(destPeer.Address) + "/" rumor.PeerMessage.ID + 1
	statusString := fmt.Sprintf("%s %s %d", rumor.Origin, addrToString(destPeer.Address), rumor.ID+1)

	// register for listener
	g.waitersMutex.Lock()
	_, present := g.gossiperWaiters[statusString]
	g.waitersMutex.Unlock()

	if present {
		// there is a goroutine already waiting for this status message

		// too bad
		return
	}

	// Register
	g.waitersMutex.Lock()
	g.gossiperWaiters[statusString] = statusChannel
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
			close(statusChannel)
			g.waitersMutex.Lock()
			g.gossiperWaiters[statusString] = nil
			g.waitersMutex.Unlock()
			// rumormonger again with probability 1/2
			if flipCoin() {
				g.standardOutputQueue <- CoinFlipString(&destPeer.Address)
				nextDestPeer := g.peerSet.RandomPeer()
				if destPeer != nil {
					go g.rumormonger(rumor, nextDestPeer)
				}
			}
			return
		case <-statusChannel:
			// received the ack before timeout
			// compare status vector
			timer.Stop()
			close(statusChannel)
			g.waitersMutex.Lock()
			g.gossiperWaiters[statusString] = nil
			g.waitersMutex.Unlock()
			return
		}
	}()

	return
}
