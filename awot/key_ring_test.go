// Tests for the Key Ring
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// peer is a network peer
type peer struct {
	id  string
	key rsa.PrivateKey
}

func TestKeyRing(t *testing.T) {
	source := "source"
	sourceKey, err := rsa.GenerateKey(rand.Reader, 512)

	if err != nil {
		t.Fatalf("could not generate rsa key: %v", err)
	}

	// generate trusted peers
	trustedPeersNames := "ABC"
	trustedPeers := make(map[string]peer)
	var trustedPeersRecords []TrustedKeyRecord

	for _, c := range trustedPeersNames {

		key, err := rsa.GenerateKey(rand.Reader, 512)

		if err != nil {
			t.Fatalf("could not generate rsa key: %v", err)
		}

		rec := TrustedKeyRecord{
			Record: KeyRecord{
				Owner:  string(c),
				KeyPub: key.PublicKey,
			},
			Confidence: 1.0,
		}
		trustedPeers[string(c)] = peer{
			string(c),
			*key,
		}
		trustedPeersRecords = append(trustedPeersRecords, rec)
	}

	ring := NewKeyRing(source, sourceKey.PublicKey, trustedPeersRecords)

	for id, _ := range trustedPeers {
		_, ok := ring.GetRecord(id)
		if !ok {
			t.Fatalf("could not get record of %s in keyring", id)
		}
	}

}

/*

	// generate keyExchangeMessages from the trusted peers to other peers
	var msgs []KeyExchangeMessage

	for _, tp := range trustedPeers {
		for oid, op := range otherPeers {
			// tid is signing for oid
			bs, err := SerializeKey(op.key.PublicKey)

			if err != nil {
				t.Fatalf("could not serialize rsa public key: %v", err)
			}

			msg := create(bs, oid, tp.key, tp.id)
			msgs = append(msgs, msg)
		}
	}

	t.Run("adding and retrieving", func(t *testing.T) {
		// Adding some unverified
		for i, msg := range msgs {
			if i%2 == 0 {
				ring.AddUnverified(msg)
			} else {
				pub, err := DeserializeKey(msg.KeyBytes)
				if err != nil {
					t.Fatalf("could not deserialize rsa public key: %v", err)
				}
				rec := KeyRecord{msg.Owner, pub}
				ring.Add(rec, msg.Origin, 0.5)
			}
		}
		time.Sleep(time.Duration(5) * time.Second)

		// getting
		tRX, ok := ring.GetRecord("X")
		if !ok {
			t.Fatalf("could not retrieve key of X")
		}

		if tRX.Confidence != 0.5 {
			t.Fatalf("confidence of X should be 0.5, got: %v", tRX.Confidence)
		}

	})

*/
