// Tests for KeyTable
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestSigningExchangeMessage(t *testing.T) {
	table := newKeyTable()

	r1K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r1 := TrustedKeyRecord{
		record: KeyRecord{
			Owner:  "node1",
			KeyPub: r1K.PublicKey,
		},
		confidence:         1.0,
		keyExchangeMessage: nil,
	}

	table.add(r1)
	rec1, _ := table.get("node1")

	table.getTrustedKeys(*r1K, "mynode")

	if rec1.keyExchangeMessage != nil {
		t.Errorf("getTrustedKeys returns a pointer")
	}

	rec2, _ := table.get("node1")

	if rec2.keyExchangeMessage == nil {
		t.Errorf("getTrustedKeys does not sign")
	}
}

func TestAddRetrieveRemove(t *testing.T) {
	table := newKeyTable()

	_, present := table.getKey("node1")
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

	table.add(r1)
	pk1, present := table.getKey("node1")

	if !present {
		t.Errorf("cannot retrieve existing key")
	}

	if r1K.PublicKey != pk1 {
		t.Errorf("keys are different")
	}

	table.remove(r1.record.Owner)

	_, present = table.getKey("node1")

	if present {
		t.Errorf("record was not deleted")
	}

}
