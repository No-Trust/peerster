// Tests for the Key Ring
package awot

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

// peer is a network peer
type peer struct {
	id  string
	key rsa.PrivateKey
}

type edge struct {
	from string
	to   string
}

// TestGraph tests the underlying graphs functions of KeyRing
func TestGraph(t *testing.T) {
	sourceKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatalf("could not generate rsa key: %v", err)
	}

	// create a test KeyRing (since there is no trusted records, it is useless)
	ring := NewKeyRing("source", sourceKey.PublicKey, nil, 0.0)

	// contains should return true for an existing node
	if !ring.contains("source") {
		t.Fatalf("KeyRing does not contain source node")
	}

	// contains should return false for a non existing nodes
	if ring.contains("non existent node") {
		t.Fatalf("KeyRing contains a non added node")
	}

	// addNode should add a node and its probability to the KeyRing
	ring.addNode("addedNode", 0.123)
	if !ring.contains("addedNode") {
		t.Fatalf("addNode should add a non existing node")
	}
	if *ring.ids["addedNode"].probability != 0.123 {
		t.Fatalf("addNode should set the probability of the added node")
	}

	// addNode should update the probability of a node if it already exists
	ring.addNode("addedNode", 0.987)
	if *ring.ids["addedNode"].probability != 0.987 {
		t.Fatalf("addNode should update the probability of an existing node")
	}

	// addEdge should add an edge and its public key to the KeyRing
	err = ring.addEdge("source", "addedNode", sourceKey.PublicKey)
	if err != nil {
		t.Fatalf("Adding a non existing edge should be possible")
	}
	if !ring.graph.HasEdgeBetween(ring.ids["source"], ring.ids["addedNode"]) {
		t.Fatalf("KeyRing does not contain added edge")
	}

	// addEdge should not add an edge with equal source and terminal
	err = ring.addEdge("addedNode", "addedNode", rsa.PublicKey{})
	if err == nil {
		t.Fatalf("Adding an edge between same vertex should not be possible")
	}

	// addEdge should not add an edge between two non-existing nodes
	err = ring.addEdge("imaginary node", "other imaginary node", rsa.PublicKey{})
	if err == nil {
		t.Fatalf("Adding an edge with non existing nodes should not be possible")
	}

	// addEdge should not add an edge with a non existing terminal
	err = ring.addEdge("source", "non existing node", rsa.PublicKey{})
	if err == nil {
		t.Fatalf("Adding an edge with non existing terminal should not be possible")
	}
}

// TestPhi tests the phi method
func TestPhi(t *testing.T) {
	tt := []struct {
		name     string
		source   string
		nodes    []string
		edges    []edge
		terminal string
		rep      float32
		phi      float32
	}{
		{
			name:   "source -> A -> B (no rep)",
			source: "source",
			nodes: []string{
				"A", "B",
			},
			edges: []edge{
				{"source", "A"},
				{"A", "B"},
			},
			terminal: "B",
			rep:      1,
			phi:      0.5,
		},
		{
			name:   "source -> A -> B -> C - > D (no rep)",
			source: "source",
			nodes: []string{
				"A", "B", "C", "D",
			},
			edges: []edge{
				{"source", "A"},
				{"A", "B"},
				{"B", "C"},
				{"C", "D"},
			},
			terminal: "D",
			rep:      1,
			phi:      0.25,
		},
		{
			name:   "source -> A -> B (rep 0.1)",
			source: "source",
			nodes: []string{
				"A", "B",
			},
			edges: []edge{
				{"source", "A"},
				{"A", "B"},
			},
			terminal: "B",
			rep:      0.1,
			phi:      0.1,
		},
		{
			name:   "source -> A -> B (no rep)",
			source: "source",
			nodes: []string{
				"A", "B",
			},
			edges: []edge{
				{"source", "A"},
				{"A", "B"},
			},
			terminal: "source",
			rep:      1,
			phi:      1,
		},
		{
			name:   "source -> A -> B (no rep)",
			source: "source",
			nodes: []string{
				"A", "B",
			},
			edges: []edge{
				{"source", "A"},
				{"A", "B"},
			},
			terminal: "C",
			rep:      1,
			phi:      0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ring := NewKeyRing(tc.source, rsa.PublicKey{}, nil, 0)
			for _, node := range tc.nodes {
				ring.addNode(node, 1)
			}

			for _, edge := range tc.edges {
				ring.addEdge(edge.from, edge.to, rsa.PublicKey{})
			}

			p := ring.phi(tc.terminal, tc.rep)
			if p != tc.phi {
				t.Fatalf("phi of %v in %v should be %v, got %v", tc.terminal, tc.name, tc.phi, p)
			}
		})
	}

}

func TestKeyRing(t *testing.T) {
	source := "source"
	sourceKey, err := rsa.GenerateKey(rand.Reader, 512)

	if err != nil {
		t.Fatalf("could not generate rsa key: %v", err)
	}

	// generate trusted peers
	trustedPeersNames := "ABC"
	trustedPeers := make(map[string]peer)
	var trustedPeersRecords []TrustedKeyRecord

	for _, c := range trustedPeersNames {

		key, err := rsa.GenerateKey(rand.Reader, 512)

		if err != nil {
			t.Fatalf("could not generate rsa key: %v", err)
		}

		rec := TrustedKeyRecord{
			KeyRecord: KeyRecord{
				Owner:  string(c),
				KeyPub: key.PublicKey,
			},
			Confidence: 1.0,
		}
		trustedPeers[string(c)] = peer{
			string(c),
			*key,
		}
		trustedPeersRecords = append(trustedPeersRecords, rec)
	}

	ring := NewKeyRing(source, sourceKey.PublicKey, trustedPeersRecords, 0.0)

	for id := range trustedPeers {
		_, ok := ring.GetRecord(id)
		if !ok {
			t.Fatalf("could not get record of %s in keyring", id)
		}
	}

}

/*

	// generate keyExchangeMessages from the trusted peers to other peers
	var msgs []KeyExchangeMessage

	for _, tp := range trustedPeers {
		for oid, op := range otherPeers {
			// tid is signing for oid
			bs, err := SerializeKey(op.key.PublicKey)

			if err != nil {
				t.Fatalf("could not serialize rsa public key: %v", err)
			}

			msg := create(bs, oid, tp.key, tp.id)
			msgs = append(msgs, msg)
		}
	}

	t.Run("adding and retrieving", func(t *testing.T) {
		// Adding some unverified
		for i, msg := range msgs {
			if i%2 == 0 {
				ring.AddUnverified(msg)
			} else {
				pub, err := DeserializeKey(msg.KeyBytes)
				if err != nil {
					t.Fatalf("could not deserialize rsa public key: %v", err)
				}
				rec := KeyRecord{msg.Owner, pub}
				ring.Add(rec, msg.Origin, 0.5)
			}
		}
		time.Sleep(time.Duration(5) * time.Second)

		// getting
		tRX, ok := ring.GetRecord("X")
		if !ok {
			t.Fatalf("could not retrieve key of X")
		}

		if tRX.Confidence != 0.5 {
			t.Fatalf("confidence of X should be 0.5, got: %v", tRX.Confidence)
		}

	})

*/
