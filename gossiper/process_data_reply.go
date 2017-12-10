// Procedure for incoming data reply from other gossipers
package main

import (
	"net"
)

// Handler for inbound data reply
func (g *Gossiper) processDataReply(reply *DataReply, remoteaddr *net.UDPAddr) {
	// check if this peer is the destination

	if reply.Destination == g.Parameters.Identifier {

		g.standardOutputQueue <- reply.DataReplyString()

		dataReplyString := string(reply.HashValue)

		g.fileWaitersMutex.Lock()
		if g.fileWaiters[dataReplyString] != nil {
			g.fileWaiters[dataReplyString] <- reply
		}
		g.fileWaitersMutex.Unlock()

		return
	}

	// this is not the destination
	// forward the packet
	if g.Parameters.NoForward {
		return
	}

	// decrement TTL, drop if less than 0
	reply.HopLimit -= 1
	if reply.HopLimit <= 0 {
		return
	}

	// get nexthop
	nexthop := g.routingTable.Get(reply.Destination)
	if nexthop != "" {
		// only forward if we have a route
		nextHopAddress := stringToUDPAddr(nexthop)

		g.gossipOutputQueue <- &Packet{
			GossipPacket: GossipPacket{
				DataReply: reply,
			},
			Destination: nextHopAddress,
		}
	}
	return

}
