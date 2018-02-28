package awot

import (
	"crypto/rsa"
	"sync"
)

// A keyTable is a public key database, a set of TrustedKeyRecord
type keyTable struct {
	db    map[string]TrustedKeyRecord // owner name -> record
	mutex *sync.Mutex
}

// getPeerList returns a slice of string with the names of peers it has a public key
func (table keyTable) getPeerList() []string {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	peers := make([]string, 0)
	for key := range table.db {
		peers = append(peers, key)
	}
	return peers
}

// add adds a record to the key table, overwrites it if it already exists
func (table *keyTable) add(rec TrustedKeyRecord) {
	table.mutex.Lock()
	table.db[rec.Owner] = rec
	table.mutex.Unlock()
}

// remove removes a record for given key owner
func (table *keyTable) remove(owner string) {
	table.mutex.Lock()
	delete(table.db, owner)
	table.mutex.Unlock()
}

// newEmptyKeyTable creates an empty keyTable
func newEmptyKeyTable() keyTable {
	return keyTable{
		db:    make(map[string]TrustedKeyRecord),
		mutex: &sync.Mutex{},
	}
}

// NewKeyTable creates a new keyTable with own's key
func newKeyTable(owner string, key rsa.PublicKey) keyTable {
	table := newEmptyKeyTable()
	table.add(TrustedKeyRecord{
		KeyRecord: KeyRecord{
			Owner:  owner,
			KeyPub: key,
		},
		Confidence: 1.0, // confidence 100%
	})

	return table
}

// get returns the record of peer with given name and true if it exists, otherwise return false
func (table keyTable) get(name string) (TrustedKeyRecord, bool) {
	table.mutex.Lock()
	r, present := table.db[name]
	table.mutex.Unlock()
	return r, present
}

// updateConfidence updates the confidence of the association key - peer, with peer's name given
// If the association does not exist yet, do nothing if the key is not present
// If a key is given, overwrites present key
func (table *keyTable) updateConfidence(name string, confidence float32, key *rsa.PublicKey) {
	table.mutex.Lock()
	r, present := table.db[name]
	if present {
		if key != nil {
			r.KeyPub = *key
		}
		r.Confidence = confidence
		table.db[name] = r
	} else if key != nil {
		table.db[name] = TrustedKeyRecord{
			KeyRecord: KeyRecord{
				Owner:  name,
				KeyPub: *key,
			},
			Confidence: confidence,
		}
	}
	table.mutex.Unlock()
}

// getKey returns the key of peer with given name and true if it exists, otherwise return false
func (table keyTable) getKey(name string) (rsa.PublicKey, bool) {
	rec, present := table.get(name)
	return rec.KeyPub, present
}

// getFullyTrustedKeys retrieves the keys with a confidence level of 100%
// If not yet signed, sign the keys
func (table *keyTable) getFullyTrustedKeys(priK rsa.PrivateKey, origin string) []TrustedKeyRecord {
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
