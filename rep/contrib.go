package rep

/*
    Imports
*/

import (
  "math/rand"
  "time"
)

/*
    Functions
*/

/**
 * Initialization function.
 */
func init() {

  rand.Seed(time.Now().UTC().UnixNano())

}

/**
 * Checks if the given peer has a contribution-based
 * reputation and if it does not, initializes it.
 */
func (table *ReputationTable) InitContribRepForPeer(peer string) {

  table.mutex.Lock()

  if _, ok := table.contribReps[peer] ; !ok {
    table.contribReps[peer] = INIT_REP
  }

  table.mutex.Unlock()

}

/**
 * Returns the contribution-based reputation of the given peer.
 */
func (table *ReputationTable) GetContribRep(peer string) (/*rep*/ float32, /*ok*/ bool) {

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
func (table *ReputationTable) ForEachContribRep(callback func(/*peer*/ string, /*rep*/ float32)) {

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
func (table *ReputationTable) DecreaseContribRep(peer string) {
  table.updateContribRep(peer, false)
}

/**
 * Updates the contribution-based reputation of a
 * given peer from which data was received.
 */
func (table *ReputationTable) IncreaseContribRep(peer string) {
  table.updateContribRep(peer, true)
}

/**
 * Updates the contribution-based reputation of a given peer
 * based on whether data was sent to or received from that peer.
 * The new reputation is computed using an exponentially weighted
 * moving average with each "new value" being either the maximum
 * or minimum possible reputation value.
 */
func (table *ReputationTable) updateContribRep(peer string, dataReceived bool) {

  table.InitContribRepForPeer(peer)

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

func (table *ReputationTable) ContribRandomPeer() string {

  var randPeer string

  var total    float32 = 0
  var counter  float32 = 0

  table.mutex.Lock()

  for _, rep := range table.contribReps {
    total += rep
  }

  random := rand.Float32() * total

  for peer, rep := range table.contribReps {

    counter += rep

    if random < counter {
      randPeer = peer
      break
    }

  }

  table.mutex.Unlock()

  return randPeer

}
