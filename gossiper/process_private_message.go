// Procedure for incoming private messages from other gossipers
package main

import (
	"github.com/No-Trust/peerster/common"
	"net"
	"log"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/rand"
)

// Handler for inbound Private Message
func (g *Gossiper) processPrivateMessage(pm *PrivateMessage, remoteaddr *net.UDPAddr) {
	// process an inbound private message
	// check if this peer is the destination

	if pm.Dest == g.Parameters.Identifier {
		// this node is the destination


		// decipher
		secret := []byte(pm.Text)
		plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, &g.key, secret, nil)
		if err != nil {
			log.Println(err)
			return
		}
		// printing
		g.standardOutputQueue <- pm.PrivateMessageString(remoteaddr)

    // If it is a request for a sig-based reputation
    // update, create one and send it as a reply
    if pm.RepSigUpdateReq {

      log := "SENDING SIG-REP UPDATE TO " + pm.Origin
      g.standardOutputQueue <- &log

      nextHop := g.routingTable.Get(pm.Origin)

      if nextHop != "" {

        g.gossipOutputQueue <- &Packet {
          GossipPacket : GossipPacket {
            Private : &PrivateMessage {
              RepUpdate : g.reputationTable.GetSigUpdate(),
            },
          },
          Destination: stringToUDPAddr(nextHop),
        }
      }

      return

    // Otherwise, if it is a sig-based reputation update,
    // forward it to reputation system instead of client
    } else if pm.RepUpdate != nil {

      log := "RECEIVED SIG-REP UPDATE FROM " + pm.Origin + "\nPRINTING OLD REPS AND NEW REPS"
      g.standardOutputQueue <- &log
      g.reputationTable.Log()

      g.reputationTable.UpdateReputations(pm.RepUpdate, pm.Origin)

      g.reputationTable.Log()

      return

    }

		// send the message to the client, if it exists
		if g.ClientAddress != nil {
			g.clientOutputQueue <- &common.Packet{
				ClientPacket: common.ClientPacket{
					NewPrivateMessage: &common.NewPrivateMessage{
						Origin: pm.Origin,
						Dest:   pm.Dest,
						Text:   string(plaintext),
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
