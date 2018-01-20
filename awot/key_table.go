// A Public Key Table
// Contains record with Public Keys for peers, with the confidence given in the association peer-public-key
package awot

import (
	"crypto/rsa"
	"sync"
)

// A KeyTable is a public key database, a set of TrustedKeyRecord
type KeyTable struct {
	db    map[string]TrustedKeyRecord // owner name -> record
	mutex *sync.Mutex
}

// getPeerList returns a slice of string with the names of peers it has a public key
func (table KeyTable) getPeerList() []string {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	peers := make([]string, 0)
	for key, _ := range table.db {
		peers = append(peers, key)
	}
	return peers
}

// add adds a record to the key table, overwrites it if it already exists
func (table *KeyTable) add(rec TrustedKeyRecord) {
	table.mutex.Lock()
	table.db[rec.Record.Owner] = rec
	table.mutex.Unlock()
}

// remove removes a record for given key owner
func (table *KeyTable) remove(owner string) {
	table.mutex.Lock()
	delete(table.db, owner)
	table.mutex.Unlock()
}

// newKeyTable creates an empty KeyTable
func newKeyTable() KeyTable {
	return KeyTable{
		db:    make(map[string]TrustedKeyRecord),
		mutex: &sync.Mutex{},
	}
}

// NewKeyTable creates a new KeyTable with own's key
func NewKeyTable(owner string, key rsa.PublicKey) KeyTable {
	table := newKeyTable()
	table.add(TrustedKeyRecord{
		Record: KeyRecord{
			Owner:  owner,
			KeyPub: key,
		},
		Confidence: 1.0, // confidence 100%
	})

	return table
}

// get returns the record of peer with given name and true if it exists, otherwise return false
func (table KeyTable) get(name string) (TrustedKeyRecord, bool) {
	table.mutex.Lock()
	r, present := table.db[name]
	table.mutex.Unlock()
	return r, present
}

// updateConfidence updates the confidence of the association key - peer, with peer's name given
// If the association does not exist yet, do nothing if the key is not present
// If a key is given, overwrites present key
func (table *KeyTable) updateConfidence(name string, confidence float32, key *rsa.PublicKey) {
	table.mutex.Lock()
	r, present := table.db[name]
	if present {
		if key != nil {
			r.Record.KeyPub = *key
		}
		r.Confidence = confidence
		table.db[name] = r
	} else if key != nil {
		table.db[name] = TrustedKeyRecord{
			Record: KeyRecord{
				Owner:  name,
				KeyPub: *key,
			},
			Confidence: confidence,
		}
	}
	table.mutex.Unlock()
}

// getKey returns the key of peer with given name and true if it exists, otherwise return false
func (table KeyTable) getKey(name string) (rsa.PublicKey, bool) {
	rec, present := table.get(name)
	return rec.Record.KeyPub, present
}

// getFullyTrustedKeys retrieves the keys with a confidence level of 100%
// If not yet signed, sign the keys
func (table *KeyTable) getFullyTrustedKeys(priK rsa.PrivateKey, origin string) []TrustedKeyRecord {
	r := make([]TrustedKeyRecord, 0)
	table.mutex.Lock()

	for i, val := range table.db {
		if val.Confidence >= 1.0 {

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
