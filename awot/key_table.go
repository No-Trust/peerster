// A Public Key Table
// Contains record with Public Keys for peers, with the confidence given in the association peer-public-key
package awot

import (
	"crypto/rsa"
	"sync"
)

// A key record, i.e. an association (public-key, owner)
type KeyRecord struct {
	Owner  string
	KeyPub rsa.PublicKey
}

// A trusted key record, i.e. an association (public-keym owner) with a confidence level
type TrustedKeyRecord struct {
	record             KeyRecord           // the record publik key - owner
	confidence         float32             // confidence level in the assocatiation owner - public key
	keyExchangeMessage *KeyExchangeMessage // the key exchange message to be advertised by the gossiper
}

// Signs a TrustedKeyRecord if not yet signed
func (rec TrustedKeyRecord) sign(priK rsa.PrivateKey, origin string) TrustedKeyRecord {
	if rec.keyExchangeMessage == nil {
		msg := create(rec.record, priK, origin)
		rec.keyExchangeMessage = &msg
	}
	return rec
}

func (rec TrustedKeyRecord) GetMessage() *KeyExchangeMessage {
	return rec.keyExchangeMessage
}

// A key database, a set of TrustedKeyRecord
type KeyTable struct {
	db    map[string]TrustedKeyRecord // owner name -> record
	mutex *sync.Mutex
}

// Add a record to the key table, overwrites it if it already exists
func (table *KeyTable) Add(rec TrustedKeyRecord) {
	table.mutex.Lock()
	table.db[rec.record.Owner] = rec
	table.mutex.Unlock()
}

// Remove a record for given key owner
func (table *KeyTable) Remove(owner string) {
	table.mutex.Lock()
	delete(table.db, owner)
	table.mutex.Unlock()
}

// Return the record of peer with given name and true if it exists, otherwise return false
func (table KeyTable) get(name string) (TrustedKeyRecord, bool) {
	table.mutex.Lock()
	r, present := table.db[name]
	table.mutex.Unlock()
	return r, present
}

// Return the key of peer with given name and true if it exists, otherwise return false
func (table KeyTable) GetKey(name string) (rsa.PublicKey, bool) {
	rec, present := table.get(name)
	return rec.record.KeyPub, present
}

// Retrieve the keys with a confidence level of >0%
// If not yet signed, sign the keys
func (table *KeyTable) GetTrustedKeys(priK rsa.PrivateKey, origin string) []TrustedKeyRecord {
	r := make([]TrustedKeyRecord, 0)
	table.mutex.Lock()

	for i, val := range table.db {
		if val.confidence > 0.0 {

			// sign key
			if val.keyExchangeMessage == nil {
				table.db[i] = val.sign(priK, origin)
			}

			r = append(r, val)
		}
	}
	table.mutex.Unlock()
	return r
}

// Create an empty KeyTable
func NewKeyTable() KeyTable {
	return KeyTable{
		db:    make(map[string]TrustedKeyRecord),
		mutex: &sync.Mutex{},
	}
}
