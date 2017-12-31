// Key exchanger functions for receiving and sending public key records
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/No-Trust/peerster/awot"
	"github.com/No-Trust/peerster/common"
	"net"
)

// Procedure for inbound KeyExchangeMessage
func (g *Gossiper) processKeyExchangeMessage(msg awot.KeyExchangeMessage, remoteaddr *net.UDPAddr) {

	// deserialize public key
	if msg.KeyBytes == nil {
		return
	}

	pemBlock, _ := pem.Decode(*msg.KeyBytes)
	if pemBlock == nil {
		return
	}
	keypub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	common.CheckError(err)
	original, ok := keypub.(*rsa.PublicKey)
	if !ok {
		return
	}

	msg.KeyRecord.KeyPub = *original

	// check the origin against the key table
	kpub, present := g.keyRing.GetKey(msg.Origin)

	if !present {
		// received a key record from a peer with no corresponding public key in memory
		// add the message to the pending list, it may be useful after getting the key
		g.standardOutputQueue <- KeyExchangeReceiveUnverifiedString(msg.KeyRecord.Owner, *remoteaddr)
		g.keyRing.AddUnverified(msg)
		return
	}

	// check validity of signature

	err = awot.Verify(msg, kpub)
	g.standardOutputQueue <- KeyExchangeReceiveString(msg.KeyRecord.Owner, *remoteaddr, err == nil)

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
	g.keyRing.Add(msg.KeyRecord, msg.Origin)
}

// Sends the self signed signature to other peers
func (g *Gossiper) SendSignatures() {
	for _, rec := range g.trustedKeys {
		// send to a random neighbor
		sendCertificate(g, rec)
	}
}

// Send a fresh key record to a random neighbor as a rumor message
func sendCertificate(g *Gossiper, rec awot.TrustedKeyRecord) {
	msg := rec.GetMessage(g.key, g.Parameters.Identifier)

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
		g.standardOutputQueue <- KeyExchangeSendString(msg.KeyRecord.Owner, destPeer.Address)
		go g.rumormonger(&rumor, destPeer)
	}
}

//
// // Send all fully trusted key records to some random neighbors each timer seconds as rumors
// // Send as many message as there are fully trusted key records
// func keyExchanger(g *Gossiper, timer uint) {
//
// 	ticker := time.NewTicker(time.Second * time.Duration(timer)) // every rate sec
// 	defer ticker.Stop()
//
// 	for _ = range ticker.C {
//
// 		// retrieve public keys and signatures
// 		records := g.keyTable.GetTrustedKeys(g.key, g.Parameters.Name)
//
// 		msgs := make([]awot.KeyExchangeMessage, len(records))
//
// 		for i, rec := range records {
// 			msgs[i] = rec.GetMessage(g.key, g.Parameters.Name)
// 		}
//
// 		for _, msg := range msgs {
// 			// for each fully trusted keys
//
// 			nextSeq := g.vectorClock.Get(g.Parameters.Identifier)
//
// 			// create rumor from message
// 			rumor := RumorMessage{
// 				Origin:      g.Parameters.Identifier,
// 				ID:          nextSeq,
// 				Text:        "",
// 				KeyExchange: &msg,
// 			}
//
// 			// update status vector
// 			g.vectorClock.Update(g.Parameters.Identifier)
//
// 			// update messages
// 			g.messages.Add(&rumor)
//
// 			// and send the rumor
// 			destPeer := g.peerSet.RandomPeer()
// 			if destPeer != nil {
// 				g.standardOutputQueue <- KeyExchangeSendString(msg.KeyRecord.Owner, destPeer.Address)
// 				go g.rumormonger(&rumor, destPeer)
// 			}
// 		}
//
// 	}
//
// }
