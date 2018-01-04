// Client API for Peerster

package common

import (
	"net"
)

type Packet struct {
	ClientPacket ClientPacket
	Destination  net.UDPAddr
}

// A message for client <-> gossiper communication
type ClientPacket struct {
	NewNode           *NewNode           // add node request from client
	RequestUpdate     *bool              // update request from client (to update messages, nodes)
	NewFile           *NewFile           // index file request from client
	FileRequest       *FileRequest       // file request from client
	PeerSlice         *PeerSlice         // list of peers from server
	ReachableNodes    *[]string          // list of reachable nodes from server
	NewMessage        *NewMessage        // message sent from client or new message received (update client)
	NewPrivateMessage *NewPrivateMessage // private message sent from client or new private message received (update client)
	Notification      *string            // notification from gossiper to the client
	KeyRingJSON				*[]byte						 // JSON format of the key ring
}

type NewMessage struct {
	SenderName string
	Text       string
}

type NewPrivateMessage struct {
	Origin string
	Dest   string
	Text   string
}

type NewNode struct {
	NewPeer Peer
}

type NewFile struct {
	Path string
}

type FileRequest struct {
	MetaHash    []byte
	Destination string
	FileName    string
}
