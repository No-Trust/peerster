// Tests for KeyTable
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestSigningExchangeMessage(t *testing.T) {
	table := NewKeyTable()

	r1K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r1 := TrustedKeyRecord{
		record: KeyRecord{
			Owner:  "node1",
			KeyPub: r1K.PublicKey,
		},
		confidence:         0.5,
		keyExchangeMessage: nil,
	}

	table.Add(r1)
	rec1, _ := table.get("node1")

	table.GetTrustedKeys(*r1K, "mynode")

	if rec1.keyExchangeMessage != nil {
		t.Errorf("GetTrustedKeys returns a pointer")
	}

	rec2, _ := table.get("node1")

	if rec2.keyExchangeMessage == nil {
		t.Errorf("GetTrustedKeys does not sign")
	}
}

func TestAddRetrieveRemove(t *testing.T) {
	table := NewKeyTable()

	_, present := table.GetKey("node1")
	if present {
		t.Errorf("retrieving unknown key returns")
	}

	r1K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r1 := TrustedKeyRecord{
		record: KeyRecord{
			Owner:  "node1",
			KeyPub: r1K.PublicKey,
		},
	}

	table.Add(r1)
	pk1, present := table.GetKey("node1")

	if !present {
		t.Errorf("cannot retrieve existing key")
	}

	if r1K.PublicKey != pk1 {
		t.Errorf("keys are different")
	}

	table.Remove(r1.record.Owner)

	_, present = table.GetKey("node1")

	if present {
		t.Errorf("record was not deleted")
	}

}
