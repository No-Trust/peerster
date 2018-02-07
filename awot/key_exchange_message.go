package awot

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
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
