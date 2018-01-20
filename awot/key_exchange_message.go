// A message with a public key and a signature
package awot

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/No-Trust/peerster/common"
)

// A KeyExchangeMessage is a signed relation (publickey - owner)
type KeyExchangeMessage struct {
	KeyBytes  []byte // serialized public key
	Owner     string // owner of the public key
	Origin    string // signer
	Signature []byte // signature of (keyPub <-> owner)
}

// Verify verifies that the received message is signed by the pretended origin
// Returns nil if valid, an error otherwise
func Verify(msg KeyExchangeMessage, OriginKeyPub rsa.PublicKey) error {
	data := append(msg.KeyBytes, []byte(msg.Owner)...)
	newhash := sha256.New()
	newhash.Write(data)
	hashed := newhash.Sum(nil)
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	return rsa.VerifyPSS(&OriginKeyPub, crypto.SHA256, hashed, msg.Signature, nil) //&opts)
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
