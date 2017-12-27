// Procedure for incoming rumor messages from other gossipers
package main

import (
	"github.com/No-Trust/peerster/awot"
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
	g.standardOutputQueue <- g.peerSet.PeersListString()
	g.standardOutputQueue <- rumor.RumorString(remoteaddr)

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

		// this is the 'expected' message

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

		// process key exchange message
		if rumor.isKeyExchange() {
			g.processKeyExchangeMessage(*rumor.KeyExchange, remoteaddr)
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

// Procedure for inbound KeyExchangeMessage
func (g *Gossiper) processKeyExchangeMessage(msg awot.KeyExchangeMessage, remoteaddr *net.UDPAddr) {
	g.standardOutputQueue <- KeyExchangeReceiveString(msg.KeyRecord.Owner, *remoteaddr)

	// check the origin against the key table
	kpub, present := g.keyTable.GetKey(msg.Origin)

	if !present {
		// received a key record from a peer with no corresponding public key in memory
		// drop the message as it cannot be verified
		return
	}

	// check validity of signature

	err := awot.Verify(msg, kpub)

	if err != nil {
		// signature does not corresponds
		// due to either malicious peer sending this advertisment, or an error due to the network
		// TODO Raja ;-)
		g.standardOutputQueue <- WrongSignatureString()
	}

	// the signature is valid

	// update key ring
	
}
