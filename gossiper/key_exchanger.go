// Key exchanger functions for receiving and sending public key records
package main

import (
	"github.com/No-Trust/peerster/awot"
	"net"
)

// Procedure for inbound KeyExchangeMessage
func (g *Gossiper) processKeyExchangeMessage(msg *awot.KeyExchangeMessage, remoteaddr *net.UDPAddr) {

	sig := msg.Signature
	nsig := make([]byte, len(sig))
	copy(nsig, sig)
	msg.Signature = nsig

	keybytes := msg.KeyBytes
	nkeybytes := make([]byte, len(keybytes))
	copy(nkeybytes, keybytes)
	msg.KeyBytes = nkeybytes

	// deserialize key
	receivedKey, err := awot.DeserializeKey(msg.KeyBytes)
	if err != nil {
		return
	}

	// create the record
	record := awot.KeyRecord{
		Owner:  msg.Owner,
		KeyPub: receivedKey,
	}

	// check the origin against the key table
	kpub, present := g.keyRing.GetKey(msg.Origin)

	if !present {
		// received a key record from a peer with no corresponding public key in memory
		// add the message to the pending list, it may be useful after getting the key
		g.standardOutputQueue <- KeyExchangeReceiveUnverifiedString(record.Owner, msg.Origin, *remoteaddr)
		g.keyRing.AddUnverified(*msg)
		return
	}

	// check validity of signature
	err = awot.Verify(*msg, kpub)
	g.standardOutputQueue <- KeyExchangeReceiveString(record.Owner, *remoteaddr, err == nil)

	if err != nil {
		// signature does not correspond
		// either :
		//	- error in network layers below (rare)
		//	- malicious sender : either true sender, or MITM
		// TODO Raja ;-)
		return
	}

	// the signature is valid

	// update key ring
	g.keyRing.Add(record, msg.Origin)

	return
}

// Send a fresh key record to a random neighbor as a rumor message
func sendCertificate(g *Gossiper, rec awot.TrustedKeyRecord) {
	msg := rec.ConstructMessage(g.key, g.Parameters.Identifier)

	nextSeq := g.vectorClock.Get(g.Parameters.Identifier)

	// create rumor from message
	rumor := RumorMessage{
		Origin:      g.Parameters.Identifier,
		ID:          nextSeq,
		Text:        "",
		KeyExchange: &msg,
	}

	// update status vector
	g.vectorClock.Update(g.Parameters.Identifier)

	// update messages
	g.messages.Add(&rumor)

	// and send the rumor
	destPeer := g.peerSet.RandomPeer()
	if destPeer != nil {
		g.standardOutputQueue <- KeyExchangeSendString(msg.Owner, destPeer.Address)
		go g.rumormonger(&rumor, destPeer)
	}
}

// Sends the self signed signature to other peers
func (g *Gossiper) SendSignatures() {
	for _, rec := range g.trustedKeys {
		// send to a random neighbor
		sendCertificate(g, rec)
	}
}
