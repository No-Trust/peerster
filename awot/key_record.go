package awot

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/No-Trust/peerster/common"
)

// A KeyRecord is an association between a public key and an owner
type KeyRecord struct {
	Owner  string
	KeyPub rsa.PublicKey
}

// A TrustedKeyRecord is a KeyRecord with a confidence level corresponding to the trust put in the KeyRecord
type TrustedKeyRecord struct {
	KeyRecord                              // the record publik key - owner
	Confidence         float32             // confidence level in the assocatiation owner - public key
	keyExchangeMessage *KeyExchangeMessage // the key exchange message to be advertised in the future
}

// ConstructMessage constructs a KeyExchangeMessage from a TrustedKeyRecord and signs it if needed with given private key and origin name
func (rec *TrustedKeyRecord) ConstructMessage(priK rsa.PrivateKey, origin string) KeyExchangeMessage {

	rec.sign(priK, origin)

	msg := rec.keyExchangeMessage

	return *msg
}

// sign signs a TrustedKeyRecord if not yet signed, using given private key and origin name
func (rec *TrustedKeyRecord) sign(priK rsa.PrivateKey, origin string) TrustedKeyRecord {
	if rec.keyExchangeMessage == nil {

		keybytes, _ := SerializeKey(rec.KeyRecord.KeyPub)

		msg := create(keybytes, rec.KeyRecord.Owner, priK, origin)

		rec.keyExchangeMessage = &msg
	}
	return *rec
}

// create creates a KeyExchangeMessage by signing the public key record using given private key and attaching given origin name to the signature
func create(keybytes []byte, owner string, ownPrivateKey rsa.PrivateKey, origin string) KeyExchangeMessage {

	data := append(keybytes, []byte(owner)...)

	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	signature, err := rsa.SignPSS(rand.Reader, &ownPrivateKey, crypto.SHA256, hashed, nil) //&opts)
	common.CheckError(err)

	msg := KeyExchangeMessage{
		KeyBytes:  keybytes,
		Owner:     owner,
		Origin:    origin,
		Signature: signature,
	}

	return msg
}
