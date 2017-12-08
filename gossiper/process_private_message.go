// Procedure for incoming private messages from other gossipers
package main

import (
	"github.com/No-Trust/peerster/common"
	"net"
)

// Handler for inbound Private Message
func (g *Gossiper) processPrivateMessage(pm *PrivateMessage, remoteaddr *net.UDPAddr) {
	// process an inbound private message
	// check if this peer is the destination

	if pm.Dest == g.Parameters.Identifier {
		// this node is the destination

		// printing
		g.standardOutputQueue <- pm.PrivateMessageString(remoteaddr)

		// send the message to the client, if it exists
		if g.ClientAddress != nil {
			g.clientOutputQueue <- &common.Packet{
				ClientPacket: common.ClientPacket{
					NewPrivateMessage: &common.NewPrivateMessage{
						Origin: pm.Origin,
						Dest:   pm.Dest,
						Text:   pm.Text,
					},
				},
				Destination: *g.ClientAddress,
			}
		}
		return
	}

	// else, forward if allowed

	if g.Parameters.NoForward {
		return
	}

	// decrement TTL, drop if less than 0
	pm.HopLimit -= 1
	if pm.HopLimit <= 0 {
		return
	}

	// get nextHop
	nextHop := g.routingTable.Get(pm.Dest)
	if nextHop != "" {
		// Only forward if we have a route
		nextHopAddress := stringToUDPAddr(nextHop)

		// sending
		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				Private: pm,
			},
			Destination: nextHopAddress,
		}
	} else {

	}

}
