// Tests for key exchange messages methods
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// Generates key for A and B
// Create KeyExchangeMessage by A singing B's public key
// And verify that the signature is correct
func TestKeyExchangeSigning(t *testing.T) {
	A := "peerA"
	B := "peerB"

	// generate A's key
	keyA, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Errorf("Could not create private key: %v", err)
	}

	// generate B's key
	keyB, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Errorf("Could not create private key: %v", err)
	}

	recordForB := KeyRecord{
		Owner:  B,
		KeyPub: keyB.PublicKey,
	}

	// A signs B's key
	keyBBytes, err := SerializeKey(recordForB.KeyPub)

	if err != nil {
		t.Errorf("Could not serialize the key: %v", err)
	}

	msg := create(keyBBytes, recordForB.Owner, *keyA, A)

	// check that the signature is correct
	err = Verify(msg, keyA.PublicKey)

	if err != nil {
		t.Errorf("Signature of message is wrong: %v", err)
	}

	trustedRecordForB := TrustedKeyRecord{
		Record:     recordForB,
		Confidence: 1.0,
	}

	msg = trustedRecordForB.ConstructMessage(*keyA, A)

	err = Verify(msg, keyA.PublicKey)

	if err != nil {
		t.Errorf("Signature of message is wrong: %v", err)
	}
}
