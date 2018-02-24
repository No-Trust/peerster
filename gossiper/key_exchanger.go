// Key exchanger functions for receiving and sending public key records
package main

import (
	"net"

	"github.com/No-Trust/peerster/awot"
	"github.com/No-Trust/peerster/common"
)

// Procedure for inbound KeyExchangeMessage
func (g *Gossiper) processKeyExchangeMessage(msg *awot.KeyExchangeMessage, repOwner float32, remoteaddr *net.UDPAddr) {
	nsig := make([]byte, len(msg.Signature))
	copy(nsig, msg.Signature)
	msg.Signature = nsig
	nkeybytes := make([]byte, len(msg.KeyBytes))
	copy(nkeybytes, msg.KeyBytes)
	msg.KeyBytes = nkeybytes

	// deserialize key
	receivedKey, err := awot.DeserializeKey([]byte(msg.KeyBytes))
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
		common.Log(KeyExchangeReceiveUnverifiedString(record.Owner, msg.Origin, *remoteaddr),
			common.LOG_MODE_REACTIVE)
		g.keyRing.AddUnverified(*msg)
		return
	}

	// check validity of signature
	err = awot.Verify(*msg, kpub)
	common.Log(KeyExchangeReceiveString(record.Owner, *remoteaddr, err == nil),
		common.LOG_MODE_REACTIVE)

	if err != nil {
		// signature does not correspond
		// either :
		//	- error in network layers below (rare)
		//	- malicious sender : either true sender, or MITM

		// Decrease sender's reputation
		// g.reputationTable.DecreaseSigRep(/* OOPS! WE NEED SENDER'S IDENTIFIER */, record.Confidence)

		return
	}

	// the signature is valid

	// Increase sender's reputation
	// g.reputationTable.IncreaseSigRep(/* OOPS! WE NEED SENDER'S IDENTIFIER */, record.Confidence)

	// update key ring
	g.keyRing.Add(record, msg.Origin, repOwner)

	return
}

// Send a fresh key record to a random neighbor as a rumor message
func sendCertificate(g *Gossiper, rec awot.TrustedKeyRecord) {
	msg := rec.ConstructMessage(g.key, g.Parameters.Identifier)
	common.Log(KeyExchangeSignString(msg.Owner, msg.Signature), common.LOG_MODE_REACTIVE)

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
		common.Log(KeyExchangeSendString(msg.Owner, destPeer.Address), common.LOG_MODE_FULL)
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
