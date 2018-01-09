package awot

import (
	"container/list"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/No-Trust/peerster/rep"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"sync"
	"time"
)

// A Node in the key ring
type Node struct {
	name        string
	id          int64
	probability *float32
}

type Edge struct {
	F, T Node
	Key  rsa.PublicKey
}

// From returns the from-node of the edge.
func (e Edge) From() graph.Node { return e.F }

// To returns the to-node of the edge.
func (e Edge) To() graph.Node { return e.T }

func (n Node) ID() int64 {
	return int64(n.id)
}

func (n Node) DOTID() string {
	p := *n.probability
	percent := int(p * 100)
	return fmt.Sprintf("%s_%d", n.name, percent)
}

// Key Ring implementation
type KeyRing struct {
	source       string
	ids          map[string]*Node // name -> Node
	graph        simple.DirectedGraph
	nextNode     int64
	keyTable     KeyTable   // for updates
	pending      *list.List // pending keyExchangeMessage
	pendingMutex *sync.Mutex
	mutex        *sync.Mutex
}

////////// Key Ring API

// Return the key of peer with given name and true if it exists, otherwise return false
func (ring KeyRing) GetKey(name string) (rsa.PublicKey, bool) {
	return ring.keyTable.getKey(name)
}

// Return the record of peer with given name and true if it exists, otherwise return false
func (ring KeyRing) GetRecord(name string) (TrustedKeyRecord, bool) {
	return ring.keyTable.get(name)
}

// GetPeerList returns the list of peer name the keyring has a public key for
func (ring KeyRing) GetPeerList() []string {
	return ring.keyTable.getPeerList()
}

// Add an exchange message that could not be verified (lack of signer's key)
func (ring *KeyRing) AddUnverified(msg KeyExchangeMessage) {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()
	ring.pending.PushBack(msg)
}

// Update the key ring with the given key and origin of the signature
// Assumes that the signature is correct
func (ring *KeyRing) Add(rec KeyRecord, sigOrigin string, reputationOwner float32) {
	// do not update if the signer is unknown
	if !ring.contains(sigOrigin) {
		return
	}

	if rec.Owner == sigOrigin {
		// self signed record
		return
	}

	// add owner of the key if not yet known, or update its probability
	ring.addNode(rec.Owner, 0.0)

	// add edge
	err := ring.addEdge(sigOrigin, rec.Owner, rec.KeyPub)

	if err != nil {
		log.Fatal("KeyRing Add : could not add edge")
	}

	// recompute its probability
	probability := ring.phi(rec.Owner, reputationOwner)

	// update probability
	ring.addNode(rec.Owner, probability)

	// save key with "unknown" confidence, that will be computed after
	// ring.keyTable.add(TrustedKeyRecord{
	// 	Record:     rec,
	// 	Confidence: float32(0.0),
	// })

	// // recompute the confidence of the keys
	ring.updateConfidence()
}

// Create a new key-ring with given fully trusted origin-public key pairs
// The key ring will update the keytable when needed
// Creating a Key Ring will also spawn a new goroutine for updating the key ring regularly
func NewKeyRing(owner string, key rsa.PublicKey, trustedRecords []TrustedKeyRecord, reptable *rep.ReputationTable) KeyRing {

	keyTable := NewKeyTable(owner, key)
	nextNode := int64(0)

	// map
	ids := make(map[string]*Node)

	// create empty graph
	graph := simple.NewDirectedGraph()

	p := float32(1.0)
	// add source to graph
	source := Node{
		name:        owner,
		id:          nextNode,
		probability: &p,
	}
	nextNode += 1
	graph.AddNode(source)
	// set id and name association in map
	ids[owner] = &source
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
		p := float32(1.0)
		node := Node{
			name:        rec.Record.Owner,
			id:          nextNode,
			probability: &p,
		}
		nextNode += 1
		graph.AddNode(node)
		// set id and name association in map
		ids[rec.Record.Owner] = &node

		// add edge from source to new node

		// add edge from source
		edge := Edge{
			F:   source,
			T:   node,
			Key: rec.Record.KeyPub,
		}
		graph.SetEdge(edge)

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
		source:       owner,
		ids:          ids,
		graph:        *graph,
		nextNode:     nextNode,
		keyTable:     keyTable,
		pending:      list.New(),
		pendingMutex: &sync.Mutex{},
		mutex:        &sync.Mutex{},
	}

	go ring.worker(reptable)

	// return
	return ring
}

////////// Key Ring Implementation

func (ring *KeyRing) worker(reptable *rep.ReputationTable) {
	// updating the ring with yet unverified pending messages
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(5)) // every 5 sec
		defer ticker.Stop()
		for _ = range ticker.C {
			//ring.updateTrust(*reptable)
			ring.updatePending(*reptable)
			ring.updateConfidence()
			//ring.Save("ring")
		}
	}()
}

func (ring *KeyRing) updateTrust(reptable *rep.ReputationTable) {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	for name, _ := range ring.ids {
		rep, present := reptable.GetSigRep(name)
		if !present {
			rep = 0.5
		}

		probability := ring.phi(name, rep)
		ring.addNode(name, probability)
	}
}

// Update the key table : computes new confidence levels for each key
func (ring *KeyRing) updateConfidence() {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	allShortest := path.DijkstraAllPaths(&ring.graph)

	source := ring.graph.Node(ring.ids[ring.source].id)

	// compute for each node
	for terminalName, terminalVertex := range ring.ids {
		terminal := ring.graph.Node(terminalVertex.id)
		// get shortest paths from source to node
		minpaths, _ := allShortest.AllBetween(source, terminal)
		minpaths, bestKey := ring.selectBestPaths(minpaths)
		probability := ring.probabilityOfMinPaths(minpaths)
		// update the key table
		ring.keyTable.updateConfidence(terminalName, probability, bestKey)
	}
}

// selectBestPaths takes some paths and return a subset of them wich all corresponds to the same end public key
// the key chosen is the one corresponding to the maximum number of paths
// Warning : thread unsafe
func (ring KeyRing) selectBestPaths(paths [][]graph.Node) ([][]graph.Node, *rsa.PublicKey) {
	if len(paths) == 0 {
		return paths, nil
	} else if len(paths) == 1 {
		p := paths[0]
		if len(p) < 2 {
			return paths, nil
		}
		s := p[len(p)-2]
		t := p[len(p)-1]

		edge := ring.graph.Edge(s, t)
		if edge == nil {
			log.Fatal("edge disapeared")
		}
		key := edge.(Edge).Key
		return paths, &key
	}

	occurences := make(map[string]int)

	// the paths ends with the terminal
	// use the last edge if exists

	for _, p := range paths {
		if len(p) < 2 {
			// siging itself should not happen
			continue
		}
		s := p[len(p)-2]
		t := p[len(p)-1]

		edge := ring.graph.Edge(s, t)
		if edge == nil {
			log.Fatal("edge disapeared")
		}
		key := edge.(Edge).Key
		occurences[key.N.String()+"-"+string(key.E)] += 1
	}

	// find max
	max := 0
	var bkey string
	for key, occ := range occurences {
		if occ > max {
			max = occ
			bkey = key
		}
	}

	bestPaths := make([][]graph.Node, 0)
	var bestKey rsa.PublicKey
	for _, p := range paths {
		if len(p) < 2 {
			continue
		}
		s := p[len(p)-2]
		t := p[len(p)-1]

		edge := ring.graph.Edge(s, t)
		if edge == nil {
			log.Fatal("edge disapeared")
		}
		key := edge.(Edge).Key
		if bkey == key.N.String()+"-"+string(key.E) {
			bestPaths = append(bestPaths, p)
			bestKey = key
		}
	}

	return bestPaths, &bestKey
}

// update key ring with given message, if update successful return true
// an update is successful if the update was performed or enough information is known to declare the message not correct / trustworthy
func (ring *KeyRing) updateMessage(msg KeyExchangeMessage, confidenceOwner float32) bool {
	receivedKey, err := DeserializeKey(msg.KeyBytes)
	if err != nil {
		return true
	}

	record := KeyRecord{
		Owner:  msg.Owner,
		KeyPub: receivedKey,
	}

	kpub, present := ring.GetKey(msg.Origin)

	if !present {
		// still do not have a public key
		return false
	}
	err = Verify(msg, kpub)

	if err == nil {
		ring.Add(record, msg.Origin, confidenceOwner)
		return true
	}
	return false
}

// Loop over the stored unverified messages and process them
func (ring *KeyRing) updatePending(reptable rep.ReputationTable) {
	ring.pendingMutex.Lock()
	defer ring.pendingMutex.Unlock()

	toRemove := list.New()
	for e := ring.pending.Front(); e != nil; e = e.Next() {

		msg := e.Value.(KeyExchangeMessage)
		reputationOwner, ok := reptable.GetSigRep(msg.Owner)
		if !ok {
			reputationOwner = 0.5
		}
		if ring.updateMessage(msg, reputationOwner) {
			toRemove.PushBack(e)
		}
	}

	for e := ring.pending.Front(); e != nil; e = e.Next() {
		ring.pending.Remove(e)
	}
}

// Compute the probability of the node, independently of its current probability
func (ring KeyRing) phi(name string, reputation float32) float32 {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	// phi = min(1/d, rep)

	destNode := ring.graph.Node(ring.ids[name].id)
	sourceNode := ring.graph.Node(ring.ids[ring.source].id)

	// compute the distance from source to destination
	shortest := path.DijkstraFrom(sourceNode, &ring.graph)
	distance := shortest.WeightTo(destNode)

	if distance == 0 {
		distance = 1
	}
	phi := math.Min(1.0/distance, float64(reputation))

	return float32(phi)
}

// Check if the node with given name exists in the key ring
// It does not check if it exists in the key table
func (ring KeyRing) contains(name string) bool {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()
	_, present := ring.ids[name]
	return present
}

// Return the id of the noden with highest id
func (ring KeyRing) lastNode() int64 {
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

	// check if already in KeyRing
	if vp, present := ring.ids[name]; present {
		// update the probability
		*(vp.probability) = probability
		return
	}

	// add to graph
	node := Node{
		id:          ring.lastNode() + 1,
		name:        name,
		probability: &probability,
	}
	ring.nextNode += 1
	ring.graph.AddNode(node)
	ring.ids[name] = &node

	return
}

// Add a directed edge from a to b
func (ring *KeyRing) addEdge(a, b string, key rsa.PublicKey) error {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	if _, aPresent := ring.ids[a]; !aPresent {
		// a is not present
		return errors.New("adding edge with source non present")
	}

	if _, bPresent := ring.ids[b]; !bPresent {
		// a is not present
		return errors.New("adding edge with destination non present")
	}

	if a == b {
		return errors.New("adding edge between same vertices")
	}

	vA := ring.ids[a]
	vB := ring.ids[b]

	ring.graph.SetEdge(Edge{F: *vA, T: *vB, Key: key})
	return nil
}

////////// Dot and JSON Formating of Key Ring

type VertexViz struct {
	Index       int64
	Name        string
	Probability float32
	Confidence  float32
}

type EdgeViz struct {
	Source string
	Target string
}

type GraphViz struct {
	Nodes []VertexViz
	Links []EdgeViz
}

func (ring KeyRing) graphViz() GraphViz {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	nodes := make([]VertexViz, 0)
	rnodes := ring.graph.Nodes()
	for _, node := range rnodes {
		n := node.(Node)
		rec, _ := ring.GetRecord(n.name)
		v := VertexViz{
			Index:       n.ID(),
			Name:        n.name,
			Probability: *n.probability,
			Confidence:  rec.Confidence,
		}
		nodes = append(nodes, v)
	}

	links := make([]EdgeViz, 0)
	redges := ring.graph.Edges()

	for _, edge := range redges {
		e := EdgeViz{
			Source: edge.From().(Node).name,
			Target: edge.To().(Node).name,
		}
		links = append(links, e)
	}

	return GraphViz{
		Nodes: nodes,
		Links: links,
	}
}

// Marshals a keyring to a json format {nodes: ..., edges: a->b}
func (ring KeyRing) JSON() ([]byte, error) {
	gviz := ring.graphViz()
	return json.Marshal(gviz)
}

// Marshals a keyring to a dot format, or nil if error
func (ring KeyRing) Dot() *[]byte {
	ring.mutex.Lock()
	defer ring.mutex.Unlock()

	title := fmt.Sprintf("Key Ring -", time.Now().UTC().Format(time.RFC3339))

	dot, err := dot.Marshal(&(ring.graph), title, "", "", false)
	if err != nil {
		return nil
	}
	return &dot
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
		return err
	}
	f := fmt.Sprintf("%s_%s.dot", path, title)
	err = ioutil.WriteFile(f, bytes, 0644)

	return err
}
