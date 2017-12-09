// Gossiper implemetation and methods
// This is the "server" part of Peerster

package main

import (
	"github.com/No-Trust/peerster/common"
	"github.com/dedis/protobuf"
	"net"
	"sync"
)

// Gossiper implementation
type Gossiper struct {
	Parameters          Parameters   // some parameters
	gossipOutputQueue   chan *Packet // sending queue to gossip connection
	ClientAddress       *net.UDPAddr // client address
	clientOutputQueue   chan *common.Packet
	peerSet             common.PeerSet              // set of peers
	vectorClock         StatusPacket                // current state of received messages
	messages            Messages                    // set of received messages
	gossiperWaiters     map[string]chan *PeerStatus // goroutines waiting for an ack
	waitersMutex        *sync.Mutex
	fileWaiters         map[string]chan *DataReply // goroutines waiting for a data reply
	fileWaitersMutex    *sync.Mutex
	standardOutputQueue chan *string  // output queue for the standard output
	routingTable        RoutingTable  // routing table
	metadataSet         MetadataSet   // file metadatas
	FileDownloads       FileDownloads // file downloads : file that are being downloaded
}

// Create a new Gossiper
func NewGossiper(parameters Parameters, peerAddrs []net.UDPAddr) *Gossiper {

	peerSet := common.NewSetFromAddrs(peerAddrs)
	channelSize := parameters.ChannelSize
	metadataSet := NewMetadataSet()
	gossiper := Gossiper{
		Parameters:          parameters,
		gossipOutputQueue:   make(chan *Packet, channelSize),
		clientOutputQueue:   make(chan *common.Packet, channelSize),
		ClientAddress:       nil,
		peerSet:             peerSet,
		vectorClock:         *NewStatusPacket(peerSet.ToPeerArray(), parameters.Identifier),
		messages:            Messages{make(map[string]map[uint32]RumorMessage), &sync.Mutex{}},
		gossiperWaiters:     make(map[string]chan *PeerStatus),
		waitersMutex:        &sync.Mutex{},
		fileWaiters:         make(map[string]chan *DataReply),
		fileWaitersMutex:    &sync.Mutex{},
		standardOutputQueue: make(chan *string, channelSize),
		routingTable:        *NewRoutingTable(parameters.Identifier, UDPAddrToString(&parameters.GossipAddr)),
		metadataSet:         metadataSet,
		FileDownloads:       *NewFileDownloads(),
	}
	return &gossiper
}

// Start the Gossiper
func (g *Gossiper) Start() {
	var wg sync.WaitGroup
	wg.Add(7)

	// Standard output writer Thread
	go fmtwriter(g.standardOutputQueue, wg)
	// Client Listener Thread
	go listener(g.Parameters.UIConn, g, handleClientMessage, wg)
	// Client Writer Thread
	go clientwriter(g.Parameters.UIConn, g.clientOutputQueue, wg)
	// Gossiper Listener Thread
	go listener(g.Parameters.GossipConn, g, handleGossiperMessage, wg)
	// Gossiper Writer Thread
	go writer(g, g.Parameters.GossipConn, g.gossipOutputQueue, wg)
	// Anti Entropy Thread
	go antiEntropy(g, g.Parameters.Etimer, wg)
	// Route Rumor Sender thread
	go routerumor(g, g.Parameters.Rtimer, wg)

	// Broadcast a route rumor message
	broadcastNewRoute(g)

	// waiting for all goroutines to terminate
	wg.Wait()
}

// Handler for gossip packets (Status, Rumor or Private messages)
func handleGossiperMessage(buf []byte, remoteaddr *net.UDPAddr, g *Gossiper) {
	// called when a message from a peer is received
	var pkt GossipPacket
	err := protobuf.Decode(buf, &pkt)
	if common.CheckRead(err) {
		return
	}

	// A is the relay peer
	A := common.Peer{
		Address:    *remoteaddr,
		Identifier: "",
	}
	g.peerSet.Add(A) // adding A to the known peers

	// demultiplex packets
	if pkt.Rumor != nil {
		// process rumor
		g.processRumor(pkt.Rumor, remoteaddr)
	}
	if pkt.Status != nil {
		// process status
		g.processStatus(pkt.Status, remoteaddr)
	}
	if pkt.Private != nil {
		// process private message
		g.processPrivateMessage(pkt.Private, remoteaddr)
	}
	if pkt.DataRequest != nil {
		// process data request
		g.processDataRequest(pkt.DataRequest, remoteaddr)
	}
	if pkt.DataReply != nil {
		// process data reply
		g.processDataReply(pkt.DataReply, remoteaddr)
	}

	return
}

// Handler for client messages
func handleClientMessage(buf []byte, remoteaddr *net.UDPAddr, g *Gossiper) {
	var pkt common.ClientPacket
	err := protobuf.Decode(buf, &pkt)

	if common.CheckRead(err) {
		return
	}

	g.ClientAddress = remoteaddr

	// demultiplex packet

	if pkt.NewNode != nil {
		// process new node
		processNewNode(pkt.NewNode, g)
	}
	if pkt.NewMessage != nil {
		// process new message
		processNewMessage(pkt.NewMessage, g, remoteaddr)
	}
	if pkt.NewPrivateMessage != nil {
		// process new private message
		processNewPrivateMessage(pkt.NewPrivateMessage, g)
	}
	if pkt.RequestUpdate != nil {
		// process update request
		processRequestUpdate(pkt.RequestUpdate, g, remoteaddr)
	}
	if pkt.NewFile != nil {
		// process new file
		processNewFile(pkt.NewFile, g)
	}
	if pkt.FileRequest != nil {
		// process file request
		processFileRequest(pkt.FileRequest, g)
	}
}
