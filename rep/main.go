package rep

/*
   Imports
*/

import (
	"strconv"
	"sync"

	"github.com/No-Trust/peerster/common"
)

/*
   Functions
*/

/**
 * Returns a new empty reputation table.
 */
func NewReputationTable(peerSet *common.PeerSet) *ReputationTable {

	// Create a new empty reputation table
	table := ReputationTable{
		sigReps:     make(ReputationMap),
		contribReps: make(ReputationMap),
		mutex:       &sync.Mutex{},
	}

	// Get a slice of the peers in the given peerset
	peers := peerSet.ToPeerArray()

	// Add each peer to the table with initial reputation
	for _, peer := range peers {

		addr := peer.Address.IP.String() + ":" + strconv.Itoa(peer.Address.Port)

		// table.sigReps[peer.Identifier]     = INIT_REP
		table.contribReps[addr] = INIT_REP

	}

	// Return the reputation table
	return &table

}

/**
 * Finds the peer with the smallest reputation in a given
 * peer->rep map and returns a pointer to that peer.
 */
func findMinRepPeer(reps ReputationMap) string {

	// Minimum reputation and corresponding peer
	var min float32 = MAX_REP
	var minPeer string = ""

	// Loop through the entries
	for peer, rep := range reps {
		// If the current reputation is smaller than the
		// minimum, update the peer with minimum reputation
		if rep <= min {
			minPeer = peer
		}
	}

	// Return the peer
	return minPeer

}

/**
 * Returns the n highest reputations, along with their
 * corresponding peers, in the given peer->rep map, and
 * for a given n.
 */
func highestReps(reps ReputationMap, n uint) ReputationMap {

	// Make a new peer->rep map to hold the highest reputations
	highestReps := make(ReputationMap)

	// A pointer to the peer with the smallest reputation among
	// the highest reputation peers at any given time.
	// With this, when a new peer with a higher reputation than
	// this one is found, all we need to do is remove this peer
	// from the map, add the newly found peer, and update this
	// pointer with the new "smallest highest" reputation peer.
	var minPeer string = ""

	// Loop through the entries
	for peer, rep := range reps {

		// If the highest reputations map is
		// not yet full, add this entry
		if uint(len(highestReps)) < n {

			highestReps[peer] = rep

			// If the highest reputations map is full,
			// find the smallest highest reputation
			if uint(len(highestReps)) == n {
				minPeer = findMinRepPeer(highestReps)
			}

			// Otherwise, if this reputation is smaller than
			// the smallest highest reputation, add the former
			// and remove the latter
		} else if rep > highestReps[minPeer] {

			delete(highestReps, minPeer)
			highestReps[peer] = rep

			// Update the smallest highest reputation
			findMinRepPeer(highestReps)

		}

	}

	// Return the highest reputations
	return highestReps

}

/**
 * Returns a new reputation table with the n most "signature-
 * reputabale" peers and the n most "contribution-reputable"
 * peers in this table, for a given n.
 */
func (table *ReputationTable) MostReputablePeers(n uint) *ReputationTable {

	table.mutex.Lock()

	// Create the new table, populating its signature-based and
	// contribution-based maps with the highest peers from this table
	highestRepTable := &ReputationTable{
		sigReps:     highestReps(table.sigReps, n),
		contribReps: highestReps(table.contribReps, n),
	}

	table.mutex.Unlock()

	// Return the new reputation table
	return highestRepTable

}
