package rep

/*
    Functions
*/

// TODO: Find a cleaner way to return the result of the map access?!

/**
 * Returns the signature-based reputation of the given peer.
 */
func (table *ReputationTable) GetSigRep(peer string) (/*rep*/ float32, /*ok*/ bool) {
    
    rep, ok := table.sigReps[peer]
    
    return rep, ok
    
}

/**
 * Returns the contribution-based reputation of the given peer.
 */
func (table *ReputationTable) GetContribRep(peer string) (/*rep*/ float32, /*ok*/ bool) {
    
    rep, ok := table.contribReps[peer]
    
    return rep, ok
    
}
