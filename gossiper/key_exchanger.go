// Key exchanger functions for receiving and sending public key records
package main

import (
	"github.com/No-Trust/peerster/awot"
	"time"
)

// TODO : length of signature ?


// Send fully trusted key records to a random neighbor each timer seconds
// Send as many message as there are fully trusted key records
func keyExchanger(g *Gossiper, timer uint) {

	ticker := time.NewTicker(time.Second * time.Duration(timer)) // every rate sec
	defer ticker.Stop()

	for _ = range ticker.C {

		// retrieve public keys
		records := g.KeyTable.GetTrustedKeys()

		msgs := make([]awot.KeyExchangeMessage, len(records))

		for i, rec := range records {
			// sign the record if not yet signed
			// rec.Sign(g.key, g.Parameters.Name)
      // msgs[i] = rec.keyExchangeMessage
      msgs[i] = rec.ExchangeMessage(g.key, g.Parameters.Name)
		}

		// send records to a random neighbor
		A := g.peerSet.RandomPeer()


    for _, msg := range msgs {
      g.standardOutputQueue <- KeyExchangeSendString(msg.KeyRecord.Owner, A.Address)

      g.gossipOutputQueue <- &Packet{
        GossipPacket: GossipPacket{
          KeyExchange: &msg,
        },
        Destination: A.Address,
      }
    }

	}

}
