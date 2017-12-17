package automated_web_of_trust

import (
    "gonum.org/v1/gonum/graph/simple"
)

// Key Ring implementation
type KeyRing struct {
  graph simple.WeightedDirectedGraph
}
