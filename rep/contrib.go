package rep

/*
    Imports
*/

import (
  "github.com/No-Trust/peerster/common"
)

/*
    Functions
*/

/**
 * Returns the contribution-based reputation of the given peer.
 */
func (table *ReputationTable) GetContribRep(peer *common.Peer) (/*rep*/ float32, /*ok*/ bool) {

  table.mutex.Lock()

  // Get the reputation from the table
  rep, ok := table.contribReps[peer]

  table.mutex.Unlock()

  // Return the reputation
  return rep, ok

}

/**
 * Performs an operation for each entry in the contribution-based
 * reputation table. The operation is defined as a callback
 * function that takes a peer and a reputation as parameters.
 */
func (table *ReputationTable) ForEachContribRep(callback func(/*peer*/ *common.Peer, /*rep*/ float32)) {

  // Loop through the entries
  for peer, rep := range table.contribReps {
    // Call the given callback for each (peer, rep) pair
    callback(peer, rep)
  }

}

/**
 * Updates the contribution-based reputation of a
 * given peer to which data was sent.
 */
func (table *ReputationTable) UpdateContribRepDataSent(peer *common.Peer) {
  table.updateContribRep(peer, false)
}

/**
 * Updates the contribution-based reputation of a
 * given peer from which data was received.
 */
func (table *ReputationTable) UpdateContribRepDataReceived(peer *common.Peer) {
  table.updateContribRep(peer, true)
}

/**
 * Updates the contribution-based reputation of a given peer
 * based on whether data was sent to or received from that peer.
 * The new reputation is computed using an exponentially weighted
 * moving average with each "new value" being either the maximum
 * or minimum possible reputation value.
 */
func (table *ReputationTable) updateContribRep(peer *common.Peer, dataReceived bool) {

  // The new value to use in the moving average formula
  var newValue float32

  // If the data was received, set the new value to the maximum
  // possible reputation value, otherwise set it to the minimum
  if dataReceived {
    newValue = MAX_REP
  } else {
    newValue = MIN_REP
  }

  table.mutex.Lock()

  // Update the peer's reputation
  table.contribReps[peer] = CONTRIB_ALPHA * newValue +
    CONTRIB_ONE_MINUS_ALPHA * table.contribReps[peer]

  table.mutex.Unlock()

}
