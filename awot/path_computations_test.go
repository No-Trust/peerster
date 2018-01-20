// Tests for paths computations
package awot

import (
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

func TestComb(t *testing.T) {
	ns := []node{}
	for i := 0; i < 20; i++ {
		ns = append(ns, node(i))
	}

	tt := []struct {
		name  string
		paths []Path
		t     int
		combs []Path
	}{
		{"no paths", nil, 0, nil},
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
