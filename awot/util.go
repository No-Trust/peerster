package awot

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"log"
)

func serializeKey(key rsa.PublicKey) []byte {

	PubASN1, err := x509.MarshalPKIXPublicKey(&key)
	if err != nil {
    log.Printf("erro : ", err)
		log.Fatal("x509 MarshalPKIXPublicKey error")
	}

	data := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: PubASN1,
	})

	return data
}

// deserialize public key
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
