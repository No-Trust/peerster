package rep

/*
    Imports
*/

import (
  "sync"
)

/*
   Type definitions
*/

type ReputationTable struct {
	sigReps     map[string]float32
	contribReps map[string]float32
	mutex       *sync.Mutex
}
