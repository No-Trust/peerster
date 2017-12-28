// A Public Key Table
// Contains record with Public Keys for peers, with the confidence given in the association peer-public-key
package awot

import (
	"crypto/rsa"
	"sync"
)

// A key database, a set of TrustedKeyRecord
type KeyTable struct {
	db    map[string]TrustedKeyRecord // owner name -> record
	mutex *sync.Mutex
}

// Add a record to the key table, overwrites it if it already exists
func (table *KeyTable) add(rec TrustedKeyRecord) {
	table.mutex.Lock()
	table.db[rec.record.Owner] = rec
	table.mutex.Unlock()
}

// Remove a record for given key owner
func (table *KeyTable) remove(owner string) {
	table.mutex.Lock()
	delete(table.db, owner)
	table.mutex.Unlock()
}

// Create a new KeyTable with own's key
func NewKeyTable(owner string, key rsa.PublicKey) KeyTable {
	table := newKeyTable()
	table.add(TrustedKeyRecord {
		record: KeyRecord {
			Owner: owner,
			KeyPub: key,
		},
		confidence: 1.0, // confidence 100%
	})

	return table
}

// Return the record of peer with given name and true if it exists, otherwise return false
func (table KeyTable) get(name string) (TrustedKeyRecord, bool) {
	table.mutex.Lock()
	r, present := table.db[name]
	table.mutex.Unlock()
	return r, present
}

// Return the key of peer with given name and true if it exists, otherwise return false
func (table KeyTable) getKey(name string) (rsa.PublicKey, bool) {
	rec, present := table.get(name)
	return rec.record.KeyPub, present
}

// Create an empty KeyTable
func newKeyTable() KeyTable {
	return KeyTable{
		db:    make(map[string]TrustedKeyRecord),
		mutex: &sync.Mutex{},
	}
}

// Retrieve the keys with a confidence level of 100%
// If not yet signed, sign the keys
func (table *KeyTable) getTrustedKeys(priK rsa.PrivateKey, origin string) []TrustedKeyRecord {
	r := make([]TrustedKeyRecord, 0)
	table.mutex.Lock()

	for i, val := range table.db {
		if val.confidence >= 1.0 {

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
