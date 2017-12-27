// Tests for the Automated Web of Trust
package awot

import (
)
//
// // creates a basic key ring and save it to disk
// func TestKeyRingSave(t *testing.T) {
// 	ring := newKeyRing("source")
//
//   ring.addNode("nodeA", 0.6)
//   ring.addNode("nodeB", 0.2)
//   ring.addNode("nodeC", 0.9)
//
//   ring.addEdge("source", "nodeB")
//   ring.addEdge("nodeA", "nodeC")
//
//   ring.addNode("nodeD", 0.9)
//   ring.addEdge("nodeC", "nodeD")
//   ring.addEdge("nodeD", "nodeB")
//
// 	err := ring.Save("keyring.dot")
// 	if err != nil {
// 		t.Errorf("Could not save keyring")
// 	}
// }
//
//
// // creates a basic key ring and check its construction
// func TestKeyRingConstruction(t *testing.T) {
//   // TODO
// }
