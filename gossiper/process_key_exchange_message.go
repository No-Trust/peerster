//// Procedure for incoming key exchange message from other gossipers
package main

import (
	"github.com/No-Trust/peerster/awot"
  "net"
)

func (g *Gossiper) processKeyExchangeMessage(msg awot.KeyExchangeMessage, remoteaddr *net.UDPAddr) {
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
  }

  // the signature is valid

  // update key ring

}
