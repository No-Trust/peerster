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
 * Returns the signature-based reputation of the given peer.
 */
func (table *ReputationTable) GetSigRep(peer *common.Peer) (/*rep*/ float32, /*ok*/ bool) {

  // Get the reputation from the table
  rep, ok := table.sigReps[peer]

  // Return the reputation
  return rep, ok

}

// TODO: Decide where to check the signature of the
//       data and what to test in this function
/**
 * Updates the signature-based reputation of a given
 * peer from which data was received.
 */
func (table *ReputationTable) UpdateSigRep(peer *common.Peer) {

  // TEMPORARY
  correctSig := true
  var confidence float32 = 1

  if correctSig {

    table.sigReps[peer] = common.Clamp(table.sigReps[peer] +
      sigRepIncreaseFactor(confidence), MIN_REP, MAX_REP)

  } else {

    table.sigReps[peer] = common.Clamp(table.sigReps[peer] *
      sigRepDecreaseFactor(confidence), MIN_REP, MAX_REP)

  }

}

/**
 * Returns the signature-based reputation increase factor
 * to be used for the given level of confidence.
 */
func sigRepIncreaseFactor(confidence float32) float32 {
  return confidence * SIG_INCREASE_LIMIT
}

/**
 * Returns the signature-based reputation decrease factor
 * to be used for the given level of confidence.
 */
func sigRepDecreaseFactor(confidence float32) float32 {
  return 1 - confidence * SIG_DECREASE_LIMIT
}
