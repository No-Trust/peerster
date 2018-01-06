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

  table.mutex.Lock()

  // Get the reputation from the table
  rep, ok := table.sigReps[peer]

  table.mutex.Unlock()

  // Return the reputation
  return rep, ok

}

/**
 * Performs an operation for each entry in the signature-based
 * reputation table. The operation is defined as a callback
 * function that takes a peer and a reputation as parameters.
 */
func (table *ReputationTable) ForEachSigRep(callback func(/*peer*/ *common.Peer, /*rep*/ float32)) {

  // Loop through the entries
  for peer, rep := range table.sigReps {
    // Call the given callback for each (peer, rep) pair
    callback(peer, rep)
  }

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

  table.mutex.Lock()

  // If the signature is correct, increase the reputation
  // of the sending peer linearly by a factor that depends
  // on the confidence level in the public key association
  if correctSig {

    table.sigReps[peer] = common.ClampFloat32(table.sigReps[peer] +
      sigRepIncreaseFactor(confidence), MIN_REP, MAX_REP)

  // Otherwise, decrease the reputation of the sending peer
  // exponentially by a factor that depends on the confidence
  // level in the public key association
  } else {

    table.sigReps[peer] = common.ClampFloat32(table.sigReps[peer] *
      sigRepDecreaseFactor(confidence), MIN_REP, MAX_REP)

  }

  table.mutex.Unlock()

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
