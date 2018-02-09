// Tests for public key serialization
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// pubKeyEquals checks for equality of given public keys and return true if they are equal
func pubKeyEquals(a rsa.PublicKey, b rsa.PublicKey) bool {
	return a.E == b.E && a.N.Cmp(b.N) == 0
}

// TestDeSerialization tests functions SerializeKey and DeserializeKey
// DeserializeKey(SerializeKey(key)) should be equal to key
func TestDeSerialization(t *testing.T) {
	keyPri, err := rsa.GenerateKey(rand.Reader, 64)

	if err != nil {
		t.Fatalf("could not generate rsa key: %v", err)
	}

	tt := []struct {
		name       string
		privateKey *rsa.PrivateKey
		publicKey  rsa.PublicKey
	}{
		{"generated key", keyPri, keyPri.PublicKey},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := SerializeKey(tc.publicKey)
			if err != nil {
				t.Fatalf("SerializeKey of %v returned error: %v", tc.name, err)
			}
			deserialized, err := DeserializeKey(serialized)
			if err != nil {
				t.Fatalf("DeserializeKey of %v returned error: %v", tc.name, err)
			}

			if !pubKeyEquals(deserialized, tc.publicKey) {
				t.Fatalf("DeserializeKey of SerializeKey of %v should be %v, got %v", tc.name, tc.publicKey, deserialized)
			}

		})
	}

}
