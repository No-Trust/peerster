package awot

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

// VertexViz is a Vertex for a visualization of a KeyRing
type VertexViz struct {
	Index       int64
	Name        string
	Probability float32
	Confidence  float32
}

// EdgeViz is a Vertex for a visualization of a KeyRing
type EdgeViz struct {
	Source      string
	Target      string
	Fingerprint string // fingerprint of the public key, in hex format
}

// GraphViz is a Graph for a visualization of a KeyRing
type GraphViz struct {
	Nodes []VertexViz
	Links []EdgeViz
}

// JSON Marshals a KeyRing to a json format {nodes: ..., edges: a->b}
func (ring KeyRing) JSON() ([]byte, error) {
	gviz := graphViz(ring)
	return json.Marshal(gviz)
}

// graphViz returns a representation of the KeyRing in GraphViz structure
func graphViz(ring KeyRing) GraphViz {
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
			Source:      edge.From().(Node).name,
			Target:      edge.To().(Node).name,
			Fingerprint: fingerprint(edge.(Edge).Key),
		}
		links = append(links, e)
	}

	return GraphViz{
		Nodes: nodes,
		Links: links,
	}
}

// fingerprint returns the hex formatted fingerprint of the given rsa public key
func fingerprint(pub rsa.PublicKey) string {
	h := md5.New()
	binary.Write(h, binary.LittleEndian, pub.E)
	binary.Write(h, binary.LittleEndian, *pub.N)
	re := h.Sum(nil)

	re2 := make([]byte, hex.EncodedLen(len(re)))
	hex.Encode(re2, re)
	s := string(re2)

	// add ":" every two char
	for i := 2; i < len(s); i += 3 {
		s = s[:i] + ":" + s[i:]
	}
	return s
}
