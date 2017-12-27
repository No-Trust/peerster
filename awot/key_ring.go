package awot

import (
	"crypto/rsa"
	"errors"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"path/filepath"
	"sync"
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
	ids      map[nodeName]Vertice
	graph    simple.DirectedGraph
	keyTable *KeyTable
	mutex    *sync.Mutex
}

// Create a new key-ring with given fully trusted origin-public key pairs
func NewKeyRing(owner string, key rsa.PublicKey, trustedRecords []KeyRecord, keyTable *KeyTable) KeyRing {
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
	keyTable.Add(TrustedKeyRecord{
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
		keyTable.Add(TrustedKeyRecord{
			record: KeyRecord{
				Owner:  rec.Owner,
				KeyPub: rec.KeyPub,
			},
			confidence: 1.0,
		})
	}

	ring := KeyRing{
		ids:      ids,
		graph:    *graph,
		keyTable: keyTable,
		mutex:    &sync.Mutex{},
	}

	ring.Save("NEW-RING.dot")

	// return
	return ring
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
