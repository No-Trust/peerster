// Procedure for incoming status packets from other gossipers
package main

import (
	"github.com/No-Trust/peerster/common"
	"log"
	"net"
	"sync"
)

// Handler for inbound Status Packet
func (g *Gossiper) processStatus(status *StatusPacket, remoteaddr *net.UDPAddr) {
	// process an inbound status

	// printing
	g.standardOutputQueue <- g.peerSet.PeersListString()
	g.standardOutputQueue <- status.StatusString(remoteaddr)

	A := common.Peer{Address: *remoteaddr, Identifier: ""} // A is the relay peer

	for _, peerstatus := range status.Want {
		// loop over each peerstatus

		ackID := AckString(*remoteaddr, peerstatus.Identifier, peerstatus.NextID)

		g.waitersMutex.Lock()
		c, present := g.gossiperWaiters[ackID]
		if present && c != nil {
			// this is an ack
			// there is a goroutine waiting for this status message
			// so send the status to the goroutine via channel
			select {
			case c <- &peerstatus:
			default:
			}
		}
		g.waitersMutex.Unlock()

	}

	// in any case, compare state and proceed accordingly
	g.compareStateAndProcess(nil, status, &A)
}

// Compare the status vector with received status packet
// and proceed with according procedure.
func (g *Gossiper) compareStateAndProcess(rumor *RumorMessage, status *StatusPacket, destPeer *common.Peer) {

	wanting := status.Want
	status.mutex = &sync.Mutex{}

	sameState := true
	behind := false

	// create a copy of the current vector clock, so that its state is fixed
	cpy := (&g.vectorClock).Copy()

	for _, peerstatus := range wanting {
		// check for me
		myState := g.vectorClock.Get(peerstatus.Identifier)

		if myState < peerstatus.NextID {
			// behind
			behind = true
			sameState = false
			break
		}
	}

	if behind {
		// send status
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				Status: cpy,
			},
			Destination: destPeer.Address,
		}
	}

	//for _, peerstatus := range cpy.Want {
	for i := 0; i < g.vectorClock.Length(); i++ {
		peerstatus := g.vectorClock.GetIndex(i)
		myState := peerstatus.NextID
		itsState := status.Get(peerstatus.Identifier)

		if myState > itsState {
			// send wanted rumor
			sameState = false

			if !g.Parameters.NoForward {

				// rumormonger with the next message to the same peer, if allowed
				nextRumor, ok := g.messages.Get(peerstatus.Identifier, itsState)
				if ok == false {
					log.Println("", myState, itsState)
					log.Fatal("vectorClock and messages have conflicting info")
				}
				go g.rumormonger(nextRumor, destPeer)
			}
		}
	}

	if sameState {
		g.standardOutputQueue <- SyncString(&destPeer.Address)
	}

	if !g.Parameters.NoForward && sameState && rumor != nil {
		// same state for this origin
		// continue with probability 1/2, if allowed
		if flipCoin() {
			g.standardOutputQueue <- CoinFlipString(&destPeer.Address)
			// continue

			randPeer := g.reputationTable.ContribRandomPeer()

			if randPeer != "" {
				go g.rumormonger(rumor, &common.Peer{
					Address: stringToUDPAddr(randPeer),
				})
			}

			return

		} else {
			// stop
			return
		}
	}

	return
}
