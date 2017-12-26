package rep

/*
    Imports
*/

import (

  "sync"

  "github.com/No-Trust/peerster/common"

)

/*
    Functions
*/

/**
 * Returns a new empty reputation table.
 */
func NewReputationTable(peerSet *common.PeerSet) ReputationTable {

  // Create a new empty reputation table
  table := ReputationTable {
    sigReps     : make(map[*common.Peer]float32),
    contribReps : make(map[*common.Peer]float32),
    mutex       : &sync.Mutex{},
  }

  // Get a slice of the peers in the given peerset
  peers := peerSet.ToPeerArray()

  // Add each peer to the table with initial reputation
  for _, peer := range peers {

    table.sigReps[&peer]     = INIT_REP
    table.contribReps[&peer] = INIT_REP

  }

  // Return the reputation table
  return table

}

// TODO: Find a cleaner way to return the result of the map access?!

/**
 * Returns the signature-based reputation of the given peer.
 */
func (table *ReputationTable) GetSigRep(peer *common.Peer) (/*rep*/ float32, /*ok*/ bool) {

  // Get the reputation from the table
  rep, ok := table.sigReps[peer]

  // Return the reputation
  return rep, ok

}

/**
 * Returns the contribution-based reputation of the given peer.
 */
func (table *ReputationTable) GetContribRep(peer *common.Peer) (/*rep*/ float32, /*ok*/ bool) {

  // Get the reputation from the table
  rep, ok := table.contribReps[peer]

  // Return the reputation
  return rep, ok

}
