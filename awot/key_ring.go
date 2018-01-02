package awot

import (
	"container/list"
	"crypto/rsa"
	"errors"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"math"
	"path/filepath"
	"sync"
	"time"
	"log"
)

type nodeName string
type nodeId int64

type Node struct {
	id nodeId
}

func (n Node) ID() int64 {
	return int64(n.id)
}

// A Node in the key ring
type Vertex struct {
	name        nodeName // name
	id          nodeId   // id
	probability float32  // probability
}

// Key Ring implementation
type KeyRing struct {
	source       nodeName
	ids          map[nodeName]Vertex
	graph        simple.DirectedGraph
	keyTable     KeyTable   // for updates
	pending      *list.List // pending keyExchangeMessage
	pendingMutex *sync.Mutex
	mutex        *sync.Mutex
}

/////					Key Ring API					/////

// Return the key of peer with given name and true if it exists, otherwise return false
func (ring KeyRing) GetKey(name string) (rsa.PublicKey, bool) {
	return ring.keyTable.getKey(name)
}

// Return the record of peer with given name and true if it exists, otherwise return false
func (ring KeyRing) GetRecord(name string) (TrustedKeyRecord, bool) {
	return ring.keyTable.get(name)
}

// Add an exchange message that could not be verified (lack of signer's key)
func (ring *KeyRing) AddUnverified(msg KeyExchangeMessage) {
	ring.mutex.Lock()
	ring.pending.PushBack(msg)
	ring.mutex.Unlock()
}

// Update the key ring with the given key and origin of the signature
// Assumes that the signature is correct
func (ring *KeyRing) Add(rec KeyRecord, sigOrigin string) {
	// do not update if the signer is unknown
	if !ring.contains(sigOrigin) {
		return
	}

	// add owner of the key if not yet known, or update its probability
	ring.addNode(rec.Owner, 0.0)

	// add edge // TODO
	err := ring.addEdge(sigOrigin, rec.Owner)

	if err != nil {
		log.Fatal("cannot add edge")
	}

	// recompute its probability
	probability := ring.phi(rec.Owner)

	// update probability
	ring.addNode(rec.Owner, probability)

	// save key with "unknown" confidence, that will be computed after
	ring.keyTable.add(TrustedKeyRecord{
		Record:     rec,
		Confidence: float32(0.0),
	})

	// // recompute the confidence of the keys
	// ring.update()
}

// Create a new key-ring with given fully trusted origin-public key pairs
// The key ring will update the keytable when needed
// Creating a Key Ring will also spawn a new goroutine for updating the key ring regularly
func NewKeyRing(owner string, key rsa.PublicKey, trustedRecords []TrustedKeyRecord) KeyRing {

	keyTable := NewKeyTable(owner, key)

	// map
	ids := make(map[nodeName]Vertex)
	// create empty graph
	graph := simple.NewDirectedGraph()

	// add source to graph
	source := graph.NewNode()
	graph.AddNode(source)
	// set id and name association in map
	sourceV := Vertex{
		name:        nodeName(owner),
		id:          nodeId(source.ID()),
		probability: 1.0,
	}
	ids[nodeName(owner)] = sourceV
	// add key
	keyTable.add(TrustedKeyRecord{
		Record: KeyRecord{
			Owner:  owner,
			KeyPub: key,
		},
		Confidence: 1.0,
	})

	// add each fully trusted key
	for _, rec := range trustedRecords {
		// add node to graph
		node := graph.NewNode()
		graph.AddNode(node)
		// set id and name association in map
		nodeV := Vertex{
			name:        nodeName(rec.Record.Owner),
			id:          nodeId(node.ID()),
			probability: 1.0,
		}
		ids[nodeName(rec.Record.Owner)] = nodeV

		// add edge from source to new node

		// add edge from source
		source := ids[nodeName(owner)]

		graph.SetEdge(simple.Edge{F: simple.Node(source.id), T: simple.Node(nodeV.id)})

		// add key
		keyTable.add(TrustedKeyRecord{
			Record: KeyRecord{
				Owner:  rec.Record.Owner,
				KeyPub: rec.Record.KeyPub,
			},
			Confidence: rec.Confidence,
		})

	}

	ring := KeyRing{
		source:       nodeName(owner),
		ids:          ids,
		graph:        *graph,
		keyTable:     keyTable,
		pending:      list.New(),
		pendingMutex: &sync.Mutex{},
		mutex:        &sync.Mutex{},
	}

	// TODO
	ring.Save("ring")

	go ring.worker()

	// return
	return ring
}

/////					Key Ring Implementation					/////

func (ring *KeyRing) worker() {

	// updating the ring with yet unverified pending messages
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(5)) // every rate 5 sec
		defer ticker.Stop()
		for _ = range ticker.C {
			ring.updatePending()
			ring.updateConfidence()
			ring.Save("ring")
		}
	}()
}

// Update the key table : computes new confidence levels for each key
func (ring *KeyRing) updateConfidence() {
	allShortest := path.DijkstraAllPaths(&ring.graph)

	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	source := ring.graph.Node(int64(ring.ids[ring.source].id))

	// compute for each node
	// for _, terminal := range ring.graph.Nodes() {
	for terminalName, terminalVertex := range ring.ids {
		terminal := ring.graph.Node(int64(terminalVertex.id))
		// get shortest paths from source to node
		minpaths, _ := allShortest.AllBetween(source, terminal)
		probability := ring.probabilityOfMinPaths(minpaths)

		// update the key table
		ring.keyTable.updateConfidence(string(terminalName), probability)
	}
}

// update key ring with given message, if update successful return true
// an update is successful if the update was performed or enough information is known to declare the message not correct / trustworthy
func (ring *KeyRing) updateMessage(msg KeyExchangeMessage) bool {
	receivedKey, err := DeserializeKey(msg.KeyBytes)
	if err != nil {
		return true
	}

	record := KeyRecord{
		Owner:  msg.Owner,
		KeyPub: receivedKey,
	}

	// fmt.Println("~~~ Update Message Pending for :", record.Owner, " signed by", msg.Origin)

	kpub, present := ring.GetKey(msg.Origin)

	//fmt.Println("~~~ key : ", kpub)
	//fmt.Println("~~~ sig:\n", msg.Signature)

	if !present {
		// still do not have a public key
		return false
	}
	err = Verify(msg, kpub)

	if err == nil {
		ring.Add(record, msg.Origin)
		// fmt.Println("~~~ Update successful ")
		return true
	}
	// fmt.Println("~~~ Update un successful ")
	return false
}

// Loop over the stored unverified messages and process them
func (ring *KeyRing) updatePending() {
	ring.pendingMutex.Lock()
	defer ring.pendingMutex.Unlock()

	i := 1
	toRemove := list.New()
	for e := ring.pending.Front(); e != nil; e = e.Next() {

		// fmt.Println("~~~ Update # ", i)
		i += 1

		// check the origin against the key table
		msg := e.Value.(KeyExchangeMessage)

		if ring.updateMessage(msg) {
			toRemove.PushBack(e)
		}
	}

	for e := ring.pending.Front(); e != nil; e = e.Next() {
		ring.pending.Remove(e)
	}
}

// Compute the probability of the node, independently of its current probability
func (ring KeyRing) phi(name string) float32 {
	// phi = min(1/d, rep)
	nodename := nodeName(name)
	ring.mutex.Lock()

	destNode := ring.graph.Node(int64(ring.ids[nodename].id))
	sourceNode := ring.graph.Node(int64(ring.ids[ring.source].id))

	// compute the distance from source to destination
	shortest := path.DijkstraFrom(sourceNode, &ring.graph)
	distance := shortest.WeightTo(destNode)

	// TODO
	// fmt.Println("--- DISTANCE from %s to %s = %f", ring.source, name, distance)

	reputation := 1.0 // TODO !!!
	if distance == 0 {
		distance = 1
	}
	phi := math.Min(1.0/distance, reputation)

	ring.mutex.Unlock()

	return float32(phi)
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

// Return the vertice associated with the given node
func (ring KeyRing) getVertex(node graph.Node) (Vertex, bool) {
	ring.mutex.Lock()

	for _, v := range ring.ids {
		if int64(v.id) == node.ID() {
			ring.mutex.Unlock()
			return v, true
		}
	}

	ring.mutex.Unlock()
	return Vertex{}, false
}

func (ring KeyRing) lastNode() int64{
	maxId := int64(0)
	nodes := ring.graph.Nodes()
	for n := range nodes {
		if int64(n) > maxId {
			maxId = int64(n)
		}
	}
	return maxId
}

// Add a Vertex to the Keyring with given name and probability
// If the Vertex already exists, update its probability
func (ring *KeyRing) addNode(name string, probability float32) {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	nodename := nodeName(name)
	// check if already in KeyRing
	if vp, present := ring.ids[nodename]; present {
		// update the probability
		vp.probability = probability
		return
	}

	//fmt.Println(ring.graph.Nodes())

	// add to graph
	//node := ring.graph.NewNode()
	node := simple.Node(ring.lastNode() + 1)
	ring.graph.AddNode(node + 1)
	ring.ids[nodename] = Vertex{
		name:        nodename,
		id:          nodeId(node.ID() + 1),
		probability: probability,
	}

	return
}

// Add a directed edge between nodes named a and b, from a to b
func (ring *KeyRing) addEdge(a, b string) error {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

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
	return nil
}

// Marshal the graph and write to file in dot format
func (ring KeyRing) Save(filename string) error {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()
	path, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	title := fmt.Sprintf("%d", time.Now().Unix())
	bytes, err := dot.Marshal(&(ring.graph), title, "", "", false)
	if err != nil {
		ring.mutex.Unlock()
		return err
	}
	f := fmt.Sprintf("%s_%s.dot", path, title)
	err = ioutil.WriteFile(f, bytes, 0644)

	return err
}
