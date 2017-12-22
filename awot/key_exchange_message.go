// A message with a public key and a signature
package awot

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"github.com/No-Trust/peerster/common"
)

type KeyExchangeMessage struct {
	KeyRecord KeyRecord // association (public-key, owner)
	Origin    string    // origin of the signature
	signature []byte    // signature of keyPub by owner
}

// Verifies that the received message is signed by the pretended origin
// Return nil if valid, an error otherwise
func verify(msg KeyExchangeMessage, OriginKeyPub rsa.PublicKey) error {
	str := fmt.Sprintf("%v", msg.KeyRecord)
	data := []byte(str)

	hash := crypto.SHA256
	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto

	return rsa.VerifyPSS(&OriginKeyPub, hash, hashed, msg.signature, &opts)
}

// Create a KeyExchangeMessage by signing the public key record, using given private key and attach own's name to the signature
// Return the new KeyExchangeMessage
func create(keyRecord KeyRecord, ownPrivateKey rsa.PrivateKey, origin string) KeyExchangeMessage {
	str := fmt.Sprintf("%v", keyRecord)
	data := []byte(str)

	hash := crypto.SHA256
	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto

	signature, err := rsa.SignPSS(rand.Reader, &ownPrivateKey, hash, hashed, &opts)
	common.CheckError(err)

	msg := KeyExchangeMessage{
		KeyRecord: keyRecord,
		Origin:    origin,
		signature: signature,
	}

	return msg
}
