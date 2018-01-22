// Generate a private key pair

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/No-Trust/peerster/awot"
	"github.com/No-Trust/peerster/common"
)

var KEY_SIZE = 4096

// Return the private key, either stored in disk with given filename, or a new one and write it to the disk under the given filename
func getKey(pubKeyFilename, filename string) rsa.PrivateKey {

	// check if file exists
	if _, err := os.Stat(filename); err == nil {
		// key exists in disk

		// decode
		var key = new(rsa.PrivateKey)
		err = loadGob(filename, key)
		common.CheckError(err)

		err = savePublicKey(key.PublicKey, pubKeyFilename)
		common.CheckError(err)

		return *key
	}

	// create the key
	key, err := rsa.GenerateKey(rand.Reader, KEY_SIZE)
	common.CheckError(err)

	// save to disk
	err = saveGob(filename, key)
	common.CheckError(err)

	err = savePublicKey(key.PublicKey, pubKeyFilename)
	common.CheckError(err)

	return *key
}

// Construct a list of keyrecords from a directory where public keys are stored
// Implicitly the name of the owner of each public key is the file name (without .pem extension)
func getPublicKeysFromDirectory(dir string, except string) []awot.TrustedKeyRecord {
	records := make([]awot.TrustedKeyRecord, 0)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return records
	}

	for _, f := range files {

		extension := filepath.Ext(f.Name())
		if extension != ".pub" {
			continue
		}

		// get name without suffix
		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		// do not add source key
		if name != except {
			// read file
			pubBytes, err := ioutil.ReadFile(dir + f.Name())
			common.CheckError(err)

			pemBlock, _ := pem.Decode(pubBytes)

			if pemBlock != nil {
				keypub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
				common.CheckError(err)

				original, ok := keypub.(*rsa.PublicKey)

				if ok {
					// good format

					// construct record using name of file as owner
					record := awot.KeyRecord{
						Owner:  name,
						KeyPub: *original,
					}

					trec := awot.TrustedKeyRecord{
						Record:     record,
						Confidence: float32(1.0),
					}

					// add record
					records = append(records, trec)
				}
			}
		}
	}

	return records
}

// Encode public key to pem and save it to file
func savePublicKey(keypub rsa.PublicKey, filename string) error {

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
func saveGob(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

// Decode Gob file
func loadGob(path string, object interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
