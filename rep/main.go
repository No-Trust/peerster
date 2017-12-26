package rep

/*
    Imports
*/

import (
  "sync"
)

/*
    Functions
*/

// TODO: Add the ability to bootstrap the table with
//       entries for peers known at startup
/**
 * Returns a new empty reputation table.
 */
func NewReputationTable() ReputationTable {

  return ReputationTable {
    sigReps     : make(map[string]float32),
    contribReps : make(map[string]float32),
    mutex       : &sync.Mutex{},
  }

}

// TODO: Find a cleaner way to return the result of the map access?!

/**
 * Returns the signature-based reputation of the given peer.
 */
func (table *ReputationTable) GetSigRep(peer string) (/*rep*/ float32, /*ok*/ bool) {

  rep, ok := table.sigReps[peer]

  return rep, ok

}

/**
 * Returns the contribution-based reputation of the given peer.
 */
func (table *ReputationTable) GetContribRep(peer string) (/*rep*/ float32, /*ok*/ bool) {

  rep, ok := table.contribReps[peer]

  return rep, ok

}
