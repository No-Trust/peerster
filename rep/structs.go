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

type ReputationTable struct {
	sigReps     map[*common.Peer]float32
	contribReps map[*common.Peer]float32
	mutex       *sync.Mutex
}

type RepUpdate struct {
  sigReps     map[*common.Peer]float32
  contribReps map[*common.Peer]float32
}
