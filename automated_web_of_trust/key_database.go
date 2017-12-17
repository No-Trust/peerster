// A Public Key Database
package automated_web_of_trust


import (
  "crypto/rsa"
)

// A trusted key record, i.e. an association (public-key, owner)
type KeyRecord struct {
  Owner string
  KeyPub rsa.PublicKey
}

type TrustedKeyRecord struct {
  record KeyRecord
  confidence float32
}

type KeyDatabase struct {
  db map[string]TrustedKeyRecord
}
