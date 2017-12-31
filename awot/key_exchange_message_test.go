// Tests for the Automated Web of Trust
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func BenchmarkSigning(t *testing.B) {
	keyA, err := rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		t.Errorf("Could not create private key")
	}

	keyB, err := rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		t.Errorf("Could not create private key")
	}

	record := KeyRecord{
		Owner:  "peerB",
		KeyPub: keyB.PublicKey,
	}

	for i := 0; i < t.N; i++ {
		create(record, *keyA, "peerA")
	}
}

// Generates key for A and B
// Create KeyExchangeMessage by A singing B's public key
// And verify that the signature is correct
func TestKeyExchangeSigning(t *testing.T) {
	A := "peerA"
	B := "peerB"

	// generate A's key
	keyA, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("Could not create private key")
	}

	// generate B's key
	keyB, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("Could not create private key")
	}

	recordForB := KeyRecord{
		Owner:  B,
		KeyPub: keyB.PublicKey,
	}

	// A signs B's key
	msg := create(recordForB, *keyA, A)

	// check that the signature is correct
	err = Verify(msg, keyA.PublicKey)

	if err != nil {
		t.Errorf("Signature of message is wrong")
	}
}
