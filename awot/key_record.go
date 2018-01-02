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
	Record             KeyRecord           // the record publik key - owner
	Confidence         float32             // confidence level in the assocatiation owner - public key
	keyExchangeMessage *KeyExchangeMessage // the key exchange message to be advertised by the gossiper
}

// Signs a TrustedKeyRecord if not yet signed
func (rec *TrustedKeyRecord) sign(priK rsa.PrivateKey, origin string) TrustedKeyRecord {
	if rec.keyExchangeMessage == nil {

		keybytes := serializeKey(rec.Record.KeyPub)

		msg := create(keybytes, rec.Record.Owner, priK, origin)
		// fmt.Println("SIGNING ", rec.Record.Owner, "using ", priK.PublicKey)

		rec.keyExchangeMessage = &msg
	}
	return *rec
}

func (rec *TrustedKeyRecord) ConstructMessage(priK rsa.PrivateKey, origin string) KeyExchangeMessage {

	// sign
	rec.sign(priK, origin)

	msg := rec.keyExchangeMessage

	return *msg
}
