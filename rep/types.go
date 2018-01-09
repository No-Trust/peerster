package rep

/*
    Imports
*/

import "sync"

/*
    Type definitions
*/

/**
 * A simple map associating reputations in the form of
 * 32-bit floating point numbers to pointers to peers.
 */
type ReputationMap map[string]float32

/**
 * A data structure assotiating signature-based and
 * contribution-based reputations to peers using
 * ReputationMap fields.
 */
type ReputationTable struct {
	sigReps     ReputationMap
	contribReps ReputationMap
	mutex       *sync.Mutex
}

/**
 * A reputation table update, holding either signature-
 * based or contribution-based reputations.
 */
type RepUpdate struct {
  SigReps     ReputationMap
  ContribReps ReputationMap
}
