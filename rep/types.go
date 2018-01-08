package rep

/*
    Imports
*/

import (

  "sync"

  "github.com/No-Trust/peerster/common"

)

/*
    Type definitions
*/

/**
 * A data structure assotiating signature-based and
 * contribution-based reputations to pointers to peers,
 * in the form of 32-bit floating point numbers.
 */
type ReputationTable struct {
	sigReps     map[*common.Peer]float32
	contribReps map[*common.Peer]float32
	mutex       *sync.Mutex
}

/**
 * A request for a reputation table update, for either
 * signature-based or contribution-based reputations.
 */
type RepUpdateRequest struct {
  SigUpdateReq     bool
  ContribUpdateReq bool
}

/**
 * A reputation table update, holding either signature-
 * based or contribution-based reputations.
 */
type RepUpdate struct {
  SigReps     map[*common.Peer]float32
  ContribReps map[*common.Peer]float32
}
