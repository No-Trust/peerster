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
	record     KeyRecord
	confidence float32
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

func (table *KeyTable) Remove(owner string) {
	table.mutex.Lock()
	delete(table.db, owner)
	table.mutex.Unlock()
}

// Return the key of peer with given name and true if it exists, otherwise return false
func (table KeyTable) Get(name string) (rsa.PublicKey, bool) {
	table.mutex.Lock()
	r, present := table.db[name]
	table.mutex.Unlock()
	return r.record.KeyPub, present
}

// Create an empty KeyTable
func NewKeyTable() KeyTable {
	return KeyTable{
		db:    make(map[string]TrustedKeyRecord),
		mutex: &sync.Mutex{},
	}
}
