// A Public Key Database
package awot

import (
	"crypto/rsa"
)

// A key record, i.e. an association (public-key, owner)
type KeyRecord struct {
	Owner  string
	KeyPub rsa.PublicKey
}

// A trusted key record, i.e. an association (public-keym owner) with a confidence level
type TrustedKeyRecord struct {
	record     KeyRecord
	confidence float32
}

// A key database, a set of TrustedKeyRecord
type KeyDatabase struct {
	db map[string]TrustedKeyRecord
}
