// Generate a private key pair

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"github.com/No-Trust/peerster/common"
	"os"
)

var KEY_SIZE = 4096

// Return the private key, either stored in disk with given filename, or a new one and write it to the disk under the given filename
func getKey(filename string) rsa.PrivateKey {

	// check if file exists
	if _, err := os.Stat(filename); err == nil {
		// key exists in disk

		// decode
		var key = new(rsa.PrivateKey)
		err = load(filename, key)
		common.CheckError(err)

		return *key
	}

	// create the key
	key, err := rsa.GenerateKey(rand.Reader, KEY_SIZE)
	common.CheckError(err)

	// save to disk
	err = save(filename, key)
	common.CheckError(err)

	return *key
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
