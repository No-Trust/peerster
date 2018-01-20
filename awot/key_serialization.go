package awot

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/No-Trust/peerster/common"
)

// SerializeKey encodes the given public key to a x509 format and serializes it to a pem format
func SerializeKey(key rsa.PublicKey) ([]byte, error) {
	PubASN1, err := x509.MarshalPKIXPublicKey(&key)
	if err != nil {
		return nil, fmt.Errorf("could not marshal public key to x509: %v", err)
	}

	data := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: PubASN1,
	})

	return data, nil
}

// DeserializeKey deserializes a pem encoded x509 public key
func DeserializeKey(bytes []byte) (rsa.PublicKey, error) {
	pemBlock, _ := pem.Decode(bytes)
	if pemBlock == nil {
		fmt.Println("pem block nil")
		return rsa.PublicKey{}, errors.New("Key bytes does not conform to pem encoding")
	}

	keypub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	common.CheckError(err)

	original, ok := keypub.(*rsa.PublicKey)
	if !ok {
		fmt.Println("not ok")
		return rsa.PublicKey{}, errors.New("Key does not conform to rsa")
	}

	return *original, nil

}
