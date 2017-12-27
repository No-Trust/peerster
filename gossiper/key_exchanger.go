// Key exchanger functions for receiving and sending public key records
package main

import (
	"github.com/No-Trust/peerster/awot"
)

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
