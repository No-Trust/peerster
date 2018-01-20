package awot

import (
	"gonum.org/v1/gonum/graph"
)

type Path = []graph.Node

// comb returns all the intersections of any t paths in given paths
func comb(paths []Path, t int) []Path {
	c := combHelper(paths, t)

	if len(c) == 1 && len(c[0]) == 0 {
		return nil
	}
	return c
}

// combHelper is a helper for comb
func combHelper(paths []Path, t int) []Path {
	if t == 0 {
		return []Path{Path{}}
	}
	if len(paths) == 0 {
		return []Path{Path{}}
	}
	if len(paths) == t {
		return []Path{intersection(paths)}
	}

	with := combHelper(paths[1:], t-1)
	for i, _ := range with {
		with[i] = intersection([][]graph.Node{paths[0], with[i]})
	}

	without := combHelper(paths[1:], t)
	return append(with, without...)
}

// Compute the probability of the given shortest paths, using the inclusion exclusion formula
func (ring KeyRing) probabilityOfMinPaths(minpaths [][]graph.Node) float32 {
	// convert minpaths to []Path
	minPaths := make([]Path, len(minpaths))
	for i, v := range minpaths {
		vp := v
		// remove last element (target)
		if len(v) > 0 {
			vp = v[:len(v)-1]
		}
		minPaths[i] = Path(vp)
	}

	p := float32(0.0)

	s := float32(1.0)
	for i := 1; i <= len(minPaths); i++ {
		// get the possible paths of intersection of i paths in the n given
		// n choose i such paths
		npaths := comb(minPaths, i)
		for _, path := range npaths {
			pathP := ring.probabilityOfPath(path)
			p += s * pathP
		}

		s = -s
	}

	return p
}

// Compute the probability of the given path
func (ring KeyRing) probabilityOfPath(path []graph.Node) float32 {
	p := float32(1.0)

	for _, node := range path {
		v := node.(Node)
		p = p * (*(v.probability))
	}

	return p
}

// Return the intersection of given paths
// e.g. A={1,2,3} B={2,4,5}, A inter B = {1,2,3,4,5}
func intersection(paths [][]graph.Node) []graph.Node {
	var nodes []graph.Node
	for _, path := range paths {
		nodes = append(nodes, path...)
	}

	encountered := make(map[int64]bool, 0)
	r := make([]graph.Node, 0)

	for _, n := range nodes {
		if !encountered[n.ID()] {
			// not yet added
			r = append(r, n)
			encountered[n.ID()] = true
		}
	}
	return r
}
