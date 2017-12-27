// Generate a private key pair

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"github.com/No-Trust/peerster/common"
	"os"
	"errors"
	"fmt"
	"io/ioutil"
)

var KEY_SIZE = 4096

// Return the private key, either stored in disk with given filename, or a new one and write it to the disk under the given filename
func getKey(pubKeyFilename, filename string) rsa.PrivateKey {

	fmt.Println("GET KEY ", filename)

	// check if file exists
	if _, err := os.Stat(filename); err == nil {
		// key exists in disk

		// decode
		var key = new(rsa.PrivateKey)
		err = load(filename, key)
		common.CheckError(err)

		err = savePublicKey(key.PublicKey, pubKeyFilename)
		common.CheckError(err)

		return *key
	}

	// create the key
	key, err := rsa.GenerateKey(rand.Reader, KEY_SIZE)
	common.CheckError(err)

	// save to disk
	err = save(filename, key)
	common.CheckError(err)

	err = savePublicKey(key.PublicKey, pubKeyFilename)
	common.CheckError(err)

	return *key
}

// Encode public key to pem and save it to file
func savePublicKey(keypub rsa.PublicKey, filename string) error {
	fmt.Println("SAVING TO ", filename)

	// check if public key exists
	if _, err := os.Stat(filename); err != nil {


		PubASN1, err := x509.MarshalPKIXPublicKey(&keypub)
		if err != nil {
			return errors.New("could not marshal public key")
		}

		pubBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: PubASN1,
		})

		ioutil.WriteFile(filename, pubBytes, 0644)
	}

	return nil
}

// Encode via Gob to file
func save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

// Decode Gob file
func load(path string, object interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
