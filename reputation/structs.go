package reputation

import (
  "sync"
)

/*
   Type definitions
*/

type ReputationTable struct {
	Reputations map[string]float32
	mutex       *sync.Mutex
}
