package awot

import (
	"crypto/rsa"
)

// A KeyRecord is an association between a public key and an owner
type KeyRecord struct {
	Owner  string
	KeyPub rsa.PublicKey
}

// A TrustedKeyRecord is a KeyRecord with a confidence level corresponding to the trust put in the KeyRecord
type TrustedKeyRecord struct {
	Record             KeyRecord           // the record publik key - owner
	Confidence         float32             // confidence level in the assocatiation owner - public key
	keyExchangeMessage *KeyExchangeMessage // the key exchange message to be advertised in the future
}

// sign signs a TrustedKeyRecord if not yet signed, using given private key and origin name
func (rec *TrustedKeyRecord) sign(priK rsa.PrivateKey, origin string) TrustedKeyRecord {
	if rec.keyExchangeMessage == nil {

		keybytes, _ := SerializeKey(rec.Record.KeyPub)

		msg := create(keybytes, rec.Record.Owner, priK, origin)

		rec.keyExchangeMessage = &msg
	}
	return *rec
}

// ConstructMessage constructs a KeyExchangeMessage from a TrustedKeyRecord and signs it if needed with given private key and origin name
func (rec *TrustedKeyRecord) ConstructMessage(priK rsa.PrivateKey, origin string) KeyExchangeMessage {

	// sign
	rec.sign(priK, origin)

	msg := rec.keyExchangeMessage

	return *msg
}
