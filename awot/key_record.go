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
	record             KeyRecord           // the record publik key - owner
	confidence         float32             // confidence level in the assocatiation owner - public key
	keyExchangeMessage *KeyExchangeMessage // the key exchange message to be advertised by the gossiper
}

// Signs a TrustedKeyRecord if not yet signed
func (rec TrustedKeyRecord) sign(priK rsa.PrivateKey, origin string) TrustedKeyRecord {
	if rec.keyExchangeMessage == nil {
		msg := create(rec.record, priK, origin)
		rec.keyExchangeMessage = &msg
	}
	return rec
}

func (rec TrustedKeyRecord) GetMessage(priK rsa.PrivateKey, origin string) KeyExchangeMessage {
	return *rec.sign(priK, origin).keyExchangeMessage
}
