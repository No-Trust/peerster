// Tests for the Automated Web of Trust
package automated_web_of_trust

import (
	"testing"
)

// creates a basic key ring and save it to disk
func TestKeyRingSave(t *testing.T) {
	ring := NewKeyRing("source")

  ring.AddNode("nodeA", 0.6)
  ring.AddNode("nodeB", 0.2)
  ring.AddNode("nodeC", 0.9)

  ring.AddEdge("source", "nodeB")
  ring.AddEdge("nodeA", "nodeC")

  ring.AddNode("nodeD", 0.9)
  ring.AddEdge("nodeC", "nodeD")
  ring.AddEdge("nodeD", "nodeB")

	err := ring.Save("keyring.dot")
	if err != nil {
		t.Errorf("Could not save keyring")
	}
}


// creates a basic key ring and check its construction
func TestKeyRingConstruction(t *testing.T) {
  // TODO
}
