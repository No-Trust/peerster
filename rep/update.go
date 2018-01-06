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
 * Returns a new reputation update with the signature-
 * based reputations in this table.
 */
func (table *ReputationTable) GetSigUpdate() *RepUpdate {

  // Create a new reputation update
  repUpdate := RepUpdate {
		SigReps : make(map[*common.Peer]float32),
	}

  table.mutex.Lock()

  // Loop through the signature-based reputations
  // in this table and copy them to the update
  for peer, rep := range table.sigReps {
    repUpdate.SigReps[peer] = rep
  }

  table.mutex.Unlock()

  // Return the reputation update
  return &repUpdate

}

/**
 * Returns a new reputation update with the contribution-
 * based reputations in this table.
 */
func (table *ReputationTable) GetContribUpdate() *RepUpdate {

  // Create a new reputation update
  repUpdate := RepUpdate {
		ContribReps : make(map[*common.Peer]float32),
	}

  table.mutex.Lock()

  // Loop through the contribution-based reputations
  // in this table and copy them to the update
  for peer, rep := range table.contribReps {
    repUpdate.ContribReps[peer] = rep
  }

  table.mutex.Unlock()

  // Return the reputation update
  return &repUpdate

}

/**
 * Updates the reputations in this table with the ones in
 * the given reputation update weighted by the reputation
 * of the updater in this table, and updates the updater's
 * reputation based on the degree of similarity between
 * their update and this table.
 */
func (table *ReputationTable) UpdateReputations(update *RepUpdate, sender *common.Peer) {

  // The peer->rep map to update and the one
  // in the update to use for updating
  var refReps    map[*common.Peer]float32
  var senderReps map[*common.Peer]float32

  // If the signature-based map in the update is
  // non-nil, then it is a signature-based update
  if update.SigReps != nil {
    refReps    = table.sigReps
    senderReps = update.SigReps
  // Otherwise, if the contribution-based map in the update
  // is non-nil, then it is a contribution-based update
  } else if update.ContribReps != nil {
    refReps    = table.contribReps
    senderReps = update.ContribReps
  // Otherwise, return as the update is invalid
  } else {
    return
  }

  table.mutex.Lock()

  // Compute the highest reputations
  highestReps := highestReps(refReps, REP_REQ_PEER_COUNT)

  table.mutex.Unlock()

  // The updater's reputation
  var updaterRep float32
  found := false

  // Loop through the highest reputations and look for the updater
  for peer, rep := range highestReps {
    if peer == sender {
      updaterRep = rep
      found = true
      break
    }
  }

  // If the updater is not among the most reputable peers,
  // then return as they are not reputable enough to have
  // their updates taken into consideration
  if !found {
    return
  }

  // Compute the update weight based on the update
  // weight limit and the updater's reputation
  updateWeight         := updaterRep * UPDATE_WEIGHT_LIMIT
  oneMinusUpdateWeight := 1 - updateWeight

  table.mutex.Lock()

  // Loop through the updater's reputations and use
  // them to update the reputations in this table
  for peer, rep := range senderReps {
    if oldRep, ok := refReps[peer] ; ok {
      refReps[peer] = updateWeight * rep + oneMinusUpdateWeight * oldRep
    }
  }

  table.mutex.Unlock()

  // Update the updater's reputation
  table.updateUpdaterReputation(update, sender)

}

/**
 * Computes a Hamming distance between two peer->rep maps
 * and returns the average distance across the dimensions.
 */
func averageHammingDistance(reps1, reps2 map[*common.Peer]float32) float32 {

  // The total sum and count of differences
  var sum   float32 = 0
  var count float32 = 0

  // Loop through the reputations in the first map
  for peer, rep1 := range reps1 {

    // If this reputation is in the second map,
    // then add the difference in reputation to
    // the total sum and increment the total count
    if rep2, ok := reps2[peer] ; ok {

      sum += common.AbsFloat32(rep2 - rep1)
      count++

    }

  }

  // If there are no common peers in the two maps,
  // consider the two maps to be identical and
  // return an average distance of zero
  if count == 0 {
    return 0
  // Otherwise return the average distance
  } else {
    return sum / count
  }

}

/**
 * Decreases the reputation of the sender of a reputation update
 * by a factor that depends on the average Hamming distance
 * between the reputations in their update and this table.
 */
func (table *ReputationTable) updateUpdaterReputation(update *RepUpdate, updater *common.Peer) {

  // The peer->rep map and the one in the update
  // to use for when computing the Hamming distance
  var refReps    map[*common.Peer]float32
  var updateReps map[*common.Peer]float32

  // If the signature-based map in the update is
  // non-nil, then it is a signature-based update
  if update.SigReps != nil {
    refReps    = table.sigReps
    updateReps = update.SigReps
  // Otherwise, if the contribution-based map in the update
  // is non-nil, then it is a contribution-based update
  } else if update.ContribReps != nil {
    refReps    = table.contribReps
    updateReps = update.ContribReps
  // Otherwise, return as the update is invalid
  } else {
    return
  }

  table.mutex.Lock()

  // Compute the average Hamming distance
  avgDist := averageHammingDistance(refReps, updateReps)

  // Update the updater's reputation
  refReps[updater] *= 1 - avgDist * UPDATER_DECREASE_LIMIT

  table.mutex.Unlock()

}
