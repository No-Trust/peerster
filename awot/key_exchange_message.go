// A message with a public key and a signature
package awot

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"github.com/No-Trust/peerster/common"
	"log"
)

type KeyExchangeMessage struct {
	KeyRecord KeyRecord // association (public-key, owner)
	KeyBytes  *[]byte
	Origin    string // origin of the signature
	Signature []byte // signature of keyPub by owner
}

// Verifies that the received message is signed by the pretended origin
// Return nil if valid, an error otherwise
func Verify(msg KeyExchangeMessage, OriginKeyPub rsa.PublicKey) error {

	data := serializeRecord(msg.KeyRecord)

	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)

	// var opts rsa.PSSOptions
	// opts.SaltLength = rsa.PSSSaltLengthAuto

	return rsa.VerifyPSS(&OriginKeyPub, crypto.SHA256, hashed, msg.Signature, nil) //&opts)
}

// Create a KeyExchangeMessage by signing the public key record, using given private key and attach own's name to the signature
// Return the new KeyExchangeMessage
func create(keyRecord KeyRecord, ownPrivateKey rsa.PrivateKey, origin string) KeyExchangeMessage {

	data := serializeRecord(keyRecord)

	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)

	// var opts rsa.PSSOptions
	// opts.SaltLength = rsa.PSSSaltLengthAuto
	signature, err := rsa.SignPSS(rand.Reader, &ownPrivateKey, crypto.SHA256, hashed, nil) //&opts)
	common.CheckError(err)

	msg := KeyExchangeMessage{
		KeyRecord: keyRecord,
		Origin:    origin,
		Signature: signature,
	}

	return msg
}

func serializeRecord(rec KeyRecord) []byte {
	PubASN1, err := x509.MarshalPKIXPublicKey(&(rec.KeyPub))
	if err != nil {
		log.Fatal("x509 MarshalPKIXPublicKey error")
	}

	data := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: PubASN1,
	})

	return append(data, []byte(" "+rec.Owner)...)
}
