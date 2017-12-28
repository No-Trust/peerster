package awot

import (
	"crypto/rsa"
	"errors"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"sync"
	"math"
)

type nodeName string
type nodeId int64

// A Node in the key ring
type Vertice struct {
	name        nodeName // name
	id          nodeId   // id
	probability float32  // probability
}

// Key Ring implementation
type KeyRing struct {
	source   nodeName
	ids      map[nodeName]Vertice
	graph    simple.DirectedGraph
	keyTable KeyTable // for updates
	mutex    *sync.Mutex
}

// Return the key of peer with given name and true if it exists, otherwise return false
func (ring KeyRing) GetKey(name string) (rsa.PublicKey, bool) {
	return ring.keyTable.getKey(name)
}

// Update the key ring with the given key and origin of the signature
// Assumes that the signature is correct
func (ring *KeyRing) Add(rec KeyRecord, sigOrigin string) {
	// sigName := nodeName(sigOrigin)
	// ownerName := nodeName(rec.Owner)

	// do not update if the signer is unknown
	if !ring.contains(sigOrigin) {
		return
	}

	// add owner of the key if not yet known, or update its probability
	ring.addNode(rec.Owner, 0.0)

	// recompute its probability
	probability := ring.phi(rec.Owner)

	// update probability
	ring.addNode(rec.Owner, probability)

  // recompute the confidence of the keys
	ring.update()

	ring.mutex.Unlock()
}

// Update the key table with newly computed confidence levels of each key
func (ring *KeyRing) update() {
	// TODO
}

// Compute the probability of the node, independent of its current probability
func (ring KeyRing) phi(name string) float32 {
	nodename := nodeName(name)
	ring.mutex.Lock()

	destNode := ring.graph.Node(int64(ring.ids[nodename].id))
	sourceNode := ring.graph.Node(int64(ring.ids[ring.source].id))

	// compute the distance from source to destination
	shortest := path.DijkstraFrom(sourceNode, &ring.graph)
	distance := shortest.WeightTo(destNode)

	// TODO
	fmt.Println("--- DISTANCE from %s to %s = %f", ring.source, name, distance)

	reputation := 1.0 // TODO !!!

	phi := math.Min(distance, reputation)
	return float32(phi)
}

// Create a new key-ring with given fully trusted origin-public key pairs
// The key ring will update the keytable when needed
func NewKeyRing(owner string, key rsa.PublicKey, trustedRecords []KeyRecord) KeyRing {

	keyTable := NewKeyTable(owner, key)

	// map
	ids := make(map[nodeName]Vertice)
	// create empty graph
	graph := simple.NewDirectedGraph()

	// add source to graph
	source := graph.NewNode()
	graph.AddNode(source)
	// set id and name association in map
	sourceV := Vertice{
		name:        nodeName(owner),
		id:          nodeId(source.ID()),
		probability: 1.0,
	}
	ids[nodeName(owner)] = sourceV
	// add key
	keyTable.add(TrustedKeyRecord{
		record: KeyRecord{
			Owner:  owner,
			KeyPub: key,
		},
		confidence: 1.0,
	})

	// add each fully trusted key
	for _, rec := range trustedRecords {
		// add node to graph
		node := graph.NewNode()
		graph.AddNode(node)
		// set id and name association in map
		nodeV := Vertice{
			name:        nodeName(rec.Owner),
			id:          nodeId(node.ID()),
			probability: 1.0,
		}
		ids[nodeName(rec.Owner)] = nodeV

		// add edge from source to new node

		// add edge from source
		source := ids[nodeName(owner)]

		graph.SetEdge(simple.Edge{F: simple.Node(source.id), T: simple.Node(nodeV.id)})

		// add key
		keyTable.add(TrustedKeyRecord{
			record: KeyRecord{
				Owner:  rec.Owner,
				KeyPub: rec.KeyPub,
			},
			confidence: 1.0,
		})
	}

	ring := KeyRing{
		source:   nodeName(owner),
		ids:      ids,
		graph:    *graph,
		keyTable: keyTable,
		mutex:    &sync.Mutex{},
	}

	// TODO
	ring.Save("NEW-RING.dot")

	// return
	return ring
}

// Check if the node with given name exists in the key ring
// It does not check if it exists in the key table
func (ring KeyRing) contains(name string) bool {
	ring.mutex.Lock()
	nodename := nodeName(name)
	_, present := ring.ids[nodename]
	ring.mutex.Unlock()
	return present
}

// Add a Vertice to the Keyring with given name and probability
// If the Vertice already exists, update its probability
func (ring *KeyRing) addNode(name string, probability float32) {
	ring.mutex.Lock()
	nodename := nodeName(name)
	// check if already in KeyRing
	if vp, present := ring.ids[nodename]; present {
		// update the probability
		vp.probability = probability
		ring.mutex.Unlock()
		return
	}

	// add to graph
	node := ring.graph.NewNode()
	ring.graph.AddNode(node)
	ring.ids[nodename] = Vertice{
		name:        nodename,
		id:          nodeId(node.ID()),
		probability: probability,
	}

	ring.mutex.Unlock()
}

// Add a directed edge between nodes named a and b, from a to b
func (ring *KeyRing) addEdge(a, b string) error {
	ring.mutex.Lock()

	if _, aPresent := ring.ids[nodeName(a)]; !aPresent {
		// a is not present
		return errors.New("adding edge with source non present")
	}

	if _, bPresent := ring.ids[nodeName(b)]; !bPresent {
		// a is not present
		return errors.New("adding edge with destination non present")
	}

	if a == b {
		return errors.New("adding edge between same vertices")
	}

	vA := ring.ids[nodeName(a)]
	vB := ring.ids[nodeName(b)]

	ring.graph.SetEdge(simple.Edge{F: simple.Node(vA.id), T: simple.Node(vB.id)})
	ring.mutex.Unlock()
	return nil
}

// Marshal the graph and write to file in dot format
func (ring KeyRing) Save(filename string) error {
	path, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	ring.mutex.Lock()
	title := "KeyRing"
	bytes, err := dot.Marshal(&(ring.graph), title, "", "", false)
	if err != nil {
		ring.mutex.Unlock()
		return err
	}
	err = ioutil.WriteFile(path, bytes, 0644)
	ring.mutex.Unlock()

	return err
}
