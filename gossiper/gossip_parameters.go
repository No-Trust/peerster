package main

import (
	"net"
)

// Parameters of a Gossiper
type Parameters struct {
	Identifier             string      // identifier of this node
	Name                   string      // name of this node
	Etimer                 uint        // rate of anti entropy
	Rtimer                 uint        // rate of route rumors
	Reptimer               uint        // rate of reputation update requests
	Hoplimit               uint32      // TTL for the sending of private messages
	NoForward              bool        // for testing : if set, does not forward any packet except route rumors
	NatTraversal           bool        // if set, activates the nat traversal option
	GossipAddr             net.UDPAddr // ip:port of the gossip connection
	GossipConn             net.UDPConn // gossip connection
	UIAddr                 net.UDPAddr // ip:port of the client connection
	UIConn                 net.UDPConn // client connection
	ChannelSize            int         // buffered channel size (higher => better performance, less memory efficient)
	ChunkSize              uint        // size of a chunk, in byte
	FilesDirectory         string      // path to store the files
	ChunksDirectory        string      // path to store the chunks
	HashLength             uint        // length of the hashes in bits
	KeyFileName            string      // filename of stored key
	PubKeyFileName         string      // filename of stored public key
	TrustedKeysDirectory   string      // directory for the fully trusted public keys
	KeyConfidenceThreshold float32     // threshold for trusted keys
}
