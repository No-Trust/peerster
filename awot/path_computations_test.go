// Tests for paths computations
package awot

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/graph"
)

// A node is a dummy node satisfying gonum.org/v1/gonum/graph Node interface
type node int64

func (n node) ID() int64 { return int64(n) }

// pathEquals tests if two paths are equal
func pathEquals(p1 Path, p2 Path) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i].ID() != p2[i].ID() {
			return false
		}
	}

	return true
}

// pathsSetEquals tests if two sets of paths are equal
func pathsSetEquals(ps1 []Path, ps2 []Path) bool {
	if len(ps1) != len(ps2) {
		return false
	}

	for _, p1 := range ps1 {
		has := false
		for _, p2 := range ps2 {
			if pathEquals(p1, p2) {
				has = true
				break
			}
		}
		if !has {
			return false
		}
	}

	return true
}

func TestIntersection(t *testing.T) {
	ns := []node{}
	for i := 0; i < 12; i++ {
		ns = append(ns, node(i))
	}

	tt := []struct {
		name         string
		paths        []Path
		intersection Path
	}{
		{"no paths", nil, nil},
		{
			"single path {{0,1,2}}",
			[]Path{
				{ns[0], ns[1], ns[2]},
			},
			Path{ns[0], ns[1], ns[2]},
		},
		{
			"two disjoint paths {{0,1,2}, {3,4,5,6}}",
			[]Path{
				{ns[0], ns[1], ns[2]},
				{ns[3], ns[4], ns[5], ns[6]},
			},
			Path{ns[0], ns[1], ns[2], ns[3], ns[4], ns[5], ns[6]},
		},
		{
			"two overlapping paths {{0,1,2}, {1,2,5}}",
			[]Path{
				{ns[0], ns[1], ns[2]},
				{ns[1], ns[2], ns[5]},
			},
			Path{ns[0], ns[1], ns[2], ns[5]},
		},
		{
			"three overlapping paths {{0}, {0,4,5}, {4,6,8}}",
			[]Path{
				{ns[0]},
				{ns[0], ns[4], ns[5]},
				{ns[4], ns[6], ns[8]},
			},
			Path{ns[0], ns[4], ns[5], ns[6], ns[8]},
		},
		{
			"three non ordered paths {{4}, {0,4,5}, {8,6,4,5}}",
			[]Path{
				{ns[4]},
				{ns[0], ns[4], ns[5]},
				{ns[8], ns[6], ns[4], ns[5]},
			},
			Path{ns[4], ns[0], ns[5], ns[8], ns[6]},
		},
		{
			"two empty paths {{}, {}}",
			[]Path{
				{},
				{},
			},
			Path{},
		},
		{
			"two paths including one empty path  {{0, 1, 2}, {}}",
			[]Path{
				{ns[0], ns[1], ns[2]},
				{},
			},
			Path{ns[0], ns[1], ns[2]},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := intersection(tc.paths)
			if !pathEquals(r, tc.intersection) {
				t.Fatalf("intersection of %v should be %v, got %v", tc.name, tc.intersection, r)
			}
		})
	}
}

func TestComb(t *testing.T) {
	ns := []node{}
	for i := 0; i < 12; i++ {
		ns = append(ns, node(i))
	}

	tt := []struct {
		name  string
		paths []Path
		t     int
		combs []Path
	}{
		{"no paths, t is 0", nil, 0, nil},
		{"no paths, t is 2", nil, 2, nil},
		{
			"t is 0",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[3], ns[4]},
			},
			0,
			nil,
		},
		{
			"t is 1, 2 paths",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[3], ns[4]},
			},
			1,
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[3], ns[4]},
			},
		},
		{
			"t is 2, 3 paths",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[3], ns[4]},
				[]graph.Node{ns[4], ns[5], ns[6]},
			},
			2,
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2], ns[3], ns[4]},
				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[5], ns[6]},
				[]graph.Node{ns[2], ns[3], ns[4], ns[5], ns[6]},
			},
		},
		{
			"t is 2, 2 paths",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[3], ns[4]},
			},
			2,
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2], ns[3], ns[4]},
			},
		},
		{
			"t is 2, 2 disjoint paths",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[3], ns[4], ns[5]},
			},
			2,
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2], ns[3], ns[4], ns[5]},
			},
		},
		{
			"t is 3, 5 paths",
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2]},
				[]graph.Node{ns[2], ns[4], ns[5]},
				[]graph.Node{ns[4]},
				[]graph.Node{ns[0], ns[11], ns[6], ns[10]},
				[]graph.Node{ns[8], ns[2]},
			},
			3,
			[]Path{
				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[5]},
				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[5], ns[11], ns[6], ns[10]},
				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[5], ns[8]},

				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[11], ns[6], ns[10]},
				[]graph.Node{ns[0], ns[1], ns[2], ns[4], ns[8]},

				[]graph.Node{ns[0], ns[1], ns[2], ns[11], ns[6], ns[10], ns[8]},

				[]graph.Node{ns[2], ns[4], ns[5], ns[0], ns[11], ns[6], ns[10]},
				[]graph.Node{ns[2], ns[4], ns[5], ns[8]},

				[]graph.Node{ns[2], ns[4], ns[5], ns[0], ns[11], ns[6], ns[10], ns[8]},

				[]graph.Node{ns[4], ns[0], ns[11], ns[6], ns[10], ns[8], ns[2]},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := comb(tc.paths, tc.t)
			if !pathsSetEquals(r, tc.combs) {
				t.Fatalf("comb of %v should be %v, got %v", tc.name, tc.combs, r)
			}
		})
	}
}

func TestProbabilityOfPath(t *testing.T) {
	ns := []Node{}
	for i := 0; i < 10; i++ {
		p := 1.0 / float32(i+1)
		ns = append(ns, Node{
			name:        fmt.Sprintf("%d", i),
			id:          int64(i),
			probability: &p,
		})
	}

	tt := []struct {
		name string
		path Path
		prob float32
	}{
		{"empty path", nil, 1},
		{"single node path {(0, p=1)}", Path{ns[0]}, 1.0},
		{"two nodes path {(0, p=1), (1, p=0.5)}", Path{ns[0], ns[1]}, 0.5},
		{"three nodes path {(3, p=1/4), (2, p=1/3), (1, p=1/2)}", Path{ns[3], ns[2], ns[1]}, 1. / 24},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := probabilityOfPath(tc.path)
			if r != tc.prob {
				t.Fatalf("probabilityOfPath %v should be %v, got %v", tc.name, tc.prob, r)
			}
		})
	}
}

func TestProbabilityOfMinPaths(t *testing.T) {

	ns := []Node{}
	for i := 0; i < 10; i++ {
		p := float32(1.0 / 2)
		if i%2 == 0 {
			p = float32(1.0)
		}
		ns = append(ns, Node{
			name:        fmt.Sprintf("%d", i),
			id:          int64(i),
			probability: &p,
		})
	}

	tt := []struct {
		name  string
		paths []Path
		prob  float32
	}{
		{"empty paths", nil, 0.},
		{"one path {{(0, p=1), (2,p=1), (3,p=1/2)}, (4,p=1)}", []Path{{ns[0], ns[2], ns[3], ns[4]}}, 1. / 2},
		{"two paths {" +
			"\t\n{(0,p=1), (2,p=1), (3,p=1/2), (4,p=1)}," +
			"\t\n{(0,p=1), (6,p=1), (7,p=1/2), (4,p=1)}" +
			"\n}\n",
			[]Path{{ns[0], ns[2], ns[3], ns[4]}, {ns[0], ns[6], ns[7], ns[4]}},
			3. / 4,
		},
		{
			"three paths {" +
				"\t\n{(0,p=1), (2,p=1), (3,p=1/2), (4,p=1)}," +
				"\t\n{(0,p=1), (2,p=1), (5,p=1/2), (4,p=1)}" +
				"\t\n{(0,p=1), (6,p=1), (5,p=1/2), (4,p=1)}" +
				"\n}\n",
			[]Path{
				{ns[0], ns[2], ns[3], ns[4]},
				{ns[0], ns[2], ns[5], ns[4]},
				{ns[0], ns[6], ns[5], ns[4]},
			},
			3. / 4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := probabilityOfMinPaths(tc.paths)
			if r != tc.prob {
				t.Fatalf("probabilityOfMinPaths %v should be %v, got %v", tc.name, tc.prob, r)
			}
		})
	}
}
