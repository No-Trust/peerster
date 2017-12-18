package reputation

/*
    Type definitions
*/

type ReputationTable {
    Reputations  map[string]float32
    mutex        *sync.Mutex
}
