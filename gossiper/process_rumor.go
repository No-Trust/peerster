// Procedure for incoming rumor messages from other gossipers
package main

import (
	"github.com/No-Trust/peerster/common"
	"net"
)

// Handler for inbound Rumor Message
func (g *Gossiper) processRumor(rumor *RumorMessage, remoteaddr *net.UDPAddr) {
	// process an inbound rumor
	var directRoute bool = false

	if rumor.LastIP == nil || rumor.LastPort == nil {
		// this is a direct route
		directRoute = true
	} else {
		// present
		if g.Parameters.NatTraversal {

			// Add the new peer
			originPeer := common.Peer{
				Address: net.UDPAddr{
					IP:   *rumor.LastIP,
					Port: *rumor.LastPort,
					Zone: "",
				},
				Identifier: "",
			}

			g.peerSet.Add(originPeer)
		}

	}
	// overwriting LastIP and LastPort
	rumor.LastIP = &(remoteaddr.IP)
	rumor.LastPort = &(remoteaddr.Port)

	// printing
	var targetLogMode string
	if rumor.Text == "" {
		targetLogMode = common.LOG_MODE_FULL
	} else {
		targetLogMode = common.LOG_MODE_REACTIVE
	}
	common.Log(*g.peerSet.PeersListString(), common.LOG_MODE_FULL)
	common.Log(*rumor.RumorString(remoteaddr), targetLogMode)

	if g.messages.Contains(rumor) {
		if g.Parameters.NatTraversal {
			// check if this is the same previous message, and that there is a direct route
			if directRoute && rumor.ID == (g.vectorClock.Get(rumor.Origin)-1) {
				// update nex hop routing table if the route is direct, and if the rumor is the same as the previous wanted rumor
				g.routingTable.AddNextHop(rumor.Origin, remoteaddr)
			}
		}

		// do nothing
		return
	}

	// this is a new message

	if rumor.ID == g.vectorClock.Get(rumor.Origin) {

		// process key exchange message
		if rumor.isKeyExchange() {
			owner := rumor.KeyExchange.Owner
			repOwner, present := g.reputationTable.GetSigRep(owner)
			if !present {
				repOwner = 0.5
			}
			g.processKeyExchangeMessage(rumor.KeyExchange, repOwner, remoteaddr)
		}

		// this is the 'expected' message

		// Increase contribution-based reputation of sender
		g.reputationTable.IncreaseContribRep(addrToString(*remoteaddr))

		// update next hop routing table, unconditionnaly because this is a new rumor
		g.routingTable.AddNextHop(rumor.Origin, remoteaddr)

		// update messages
		g.messages.Add(rumor)

		// update status vector
		g.vectorClock.Update(rumor.Origin)

		// send to Client if Text is not empty

		if g.ClientAddress != nil && rumor.Text != "" {
			g.clientOutputQueue <- &common.Packet{
				ClientPacket: common.ClientPacket{
					NewMessage: &common.NewMessage{
						SenderName: rumor.Origin,
						Text:       rumor.Text,
					},
				},
				Destination: *g.ClientAddress,
			}
		}

	}

	// send ack
	cpy := (&g.vectorClock).Copy()
	g.gossipOutputQueue <- &Packet{
		GossipPacket: GossipPacket{
			Status: cpy,
		},
		Destination: *remoteaddr,
	}

	// then forward if allowed

	if !rumor.isRoute() && g.Parameters.NoForward {
		// disable forwarding for non route rumor message if NoForward is set
		return
	}

	fromPeer := &common.Peer{
		Address:    *remoteaddr,
		Identifier: "",
	}

	otherPeers := g.peerSet.Remove(fromPeer)

	if rumor.isRoute() {
		// forward to all peers
		go rumor.broadcastTo(g, otherPeers)

		// destPeer := (&otherPeers).RandomPeer()
		// if destPeer != nil {
		// 	go g.rumormonger(rumor, destPeer)
		// }

	} else {
		// start rumormongering
		// select a random peer

		destPeer := (&otherPeers).RandomPeer()
		if destPeer != nil {
			go g.rumormonger(rumor, destPeer)
		}
	}

	return
}
