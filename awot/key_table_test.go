// Tests for KeyTable
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// check equality of slices of strings
func equals(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func TestSigningExchangeMessage(t *testing.T) {
	table := newEmptyKeyTable()

	r1K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r1 := TrustedKeyRecord{
		KeyRecord: KeyRecord{
			Owner:  "node1",
			KeyPub: r1K.PublicKey,
		},
		Confidence: 1.0,
	}

	table.add(r1)
	rec1, _ := table.get("node1")

	table.getFullyTrustedKeys(*r1K, "mynode")

	if rec1.keyExchangeMessage != nil {
		t.Errorf("getTrustedKeys returns a pointer")
	}

	rec2, _ := table.get("node1")

	if rec2.keyExchangeMessage == nil {
		t.Errorf("getTrustedKeys does not sign")
	}
}

func TestKeyTable(t *testing.T) {
	table := newEmptyKeyTable()

	_, present := table.getKey("node1")
	if present {
		t.Errorf("retrieving unknown key returns")
	}

	r1K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r2K, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Errorf("error generating key")
	}

	r3K := r1K

	r1 := TrustedKeyRecord{
		KeyRecord: KeyRecord{
			Owner:  "node1",
			KeyPub: r1K.PublicKey,
		},
		Confidence: 0.55,
	}

	r2 := TrustedKeyRecord{
		KeyRecord: KeyRecord{
			Owner:  "node2",
			KeyPub: r2K.PublicKey,
		},
		Confidence: 0.44,
	}

	table.add(r1)
	table.add(r2)
	pk1, present := table.getKey("node1")

	if !present {
		t.Errorf("cannot retrieve existing key")
	}

	if r1K.PublicKey != pk1 {
		t.Errorf("keys are different")
	}

	retrieved, present := table.get("node1")

	if !present {
		t.Errorf("record is not present in table")
	} else if retrieved.Confidence != 0.55 {
		t.Errorf("incorrect confidence : confidence has not been stored correctly")
	}

	// checking getPeerList()
	peers := table.getPeerList()
	if !equals(peers, []string{"node1", "node2"}) && !equals(peers, []string{"node2", "node1"}) {
		t.Errorf("getPeerList incorrect")
	}

	// adding a peer with updateConfidence()
	table.updateConfidence("node3", 0.11, &r3K.PublicKey)
	retrieved, present = table.get("node3")
	if !present {
		t.Errorf("record is not present in table")
	} else if retrieved.Confidence != 0.11 {
		t.Errorf("incorrect confidence : confidence has not been stored correctly")
	}

	// updating confidence of node1 without updating the key
	table.updateConfidence("node1", 0.77, nil)

	retrieved, present = table.get("node1")

	if !present {
		t.Errorf("record is not present in table")
	} else if retrieved.Confidence != 0.77 {
		t.Errorf("incorrect confidence : confidence has not been updated with updateConfidence")
	}

	// updating confidence of node2 while updating the key
	table.updateConfidence("node1", 0.88, &r2K.PublicKey)

	retrieved, _ = table.get("node1")
	key, present := table.getKey("node1")

	if !present {
		t.Errorf("record is not present in table")
	} else if retrieved.Confidence != 0.88 {
		t.Errorf("incorrect confidence : confidence has not been updated with updateConfidence")
	}

	if key != r2K.PublicKey {
		t.Errorf("incorrect public key : public key has not been updated with updateConfidence()")
	}

	table.remove(r1.Owner)

	_, present = table.getKey("node1")

	if present {
		t.Errorf("record was not deleted from table")
	}

	retrieved, present = table.get("node2")

	if !present {
		t.Errorf("record is not present in table")
	} else if retrieved.Confidence != 0.44 {
		t.Errorf("incorrect confidence : confidence has not stored correctly")
	}

}
