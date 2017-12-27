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

  // Get the reputation from the table
  rep, ok := table.contribReps[peer]

  // Return the reputation
  return rep, ok

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
 */
func (table *ReputationTable) updateContribRep(peer *common.Peer, dataReceived bool) {

  var newValue float32

  if dataReceived {
    newValue = MAX_REP
  } else {
    newValue = MIN_REP
  }

  table.contribReps[peer] = CONTRIB_ALPHA * newValue +
    CONTRIB_ONE_MINUS_ALPHA * table.contribReps[peer]

}
