package awot

import (
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "log"
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
		msg := create(rec.Record, priK, origin)
		rec.keyExchangeMessage = &msg
	}
	return *rec
}

func (rec *TrustedKeyRecord) GetMessage(priK rsa.PrivateKey, origin string) KeyExchangeMessage {
	//return *(rec.sign(priK, origin)).keyExchangeMessage

  // sign
  rec.sign(priK, origin)

  msg := rec.keyExchangeMessage


  ////////////
  PubASN1, err := x509.MarshalPKIXPublicKey(&(rec.Record.KeyPub))
  if err != nil {
    log.Fatal("x509 MarshalPKIXPublicKey error")
  }

  pubBytes := pem.EncodeToMemory(&pem.Block{
      Type:  "RSA PUBLIC KEY",
      Bytes: PubASN1,
  })

  nrec := KeyExchangeMessage {
    KeyRecord: msg.KeyRecord,
    KeyBytes: &pubBytes,
    Origin: msg.Origin,
    Signature: msg.Signature,
  }

  return nrec
}
