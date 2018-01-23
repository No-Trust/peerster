// Gossiper implemetation and methods
// This is the "server" part of Peerster

package main

import (
	"crypto/rsa"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/No-Trust/peerster/awot"
	"github.com/No-Trust/peerster/common"
	"github.com/No-Trust/peerster/rep"
	"github.com/dedis/protobuf"
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
	standardOutputQueue chan *string            // output queue for the standard output
	routingTable        RoutingTable            // routing table
	metadataSet         MetadataSet             // file metadatas
	FileDownloads       FileDownloads           // file downloads : file that are being downloaded
	key                 rsa.PrivateKey          // private key / public key of this gossiper
	reputationTable     rep.ReputationTable     // Reputation table
	trustedKeys         []awot.TrustedKeyRecord // fully trusted keys, bootstrap of awot
	keyRing             awot.KeyRing            // key ring of awot
}

// Create a new Gossiper
func NewGossiper(parameters Parameters, peerAddrs []net.UDPAddr) *Gossiper {

	peerSet := common.NewSetFromAddrs(peerAddrs, parameters.GossipAddr)
	channelSize := parameters.ChannelSize
	metadataSet := NewMetadataSet()
	key := getKey(parameters.PubKeyFileName, parameters.KeyFileName)
	trustedKeys := getPublicKeysFromDirectory(parameters.TrustedKeysDirectory, parameters.Identifier)
	reptable := *rep.NewReputationTable(&peerSet)
	gossiper := Gossiper{
		Parameters:          parameters,
		gossipOutputQueue:   make(chan *Packet, channelSize),
		clientOutputQueue:   make(chan *common.Packet, channelSize),
		ClientAddress:       nil,
		peerSet:             peerSet,
		vectorClock:         *NewStatusPacket(peerSet.ToPeerArray(), parameters.Identifier),
		messages:            Messages{make(map[string]map[uint32]RumorMessage), &sync.Mutex{}},
		gossiperWaiters:     make(map[string]chan *PeerStatus, channelSize),
		waitersMutex:        &sync.Mutex{},
		fileWaiters:         make(map[string]chan *DataReply),
		fileWaitersMutex:    &sync.Mutex{},
		standardOutputQueue: make(chan *string, channelSize),
		routingTable:        *NewRoutingTable(parameters.Identifier, UDPAddrToString(parameters.GossipAddr)),
		metadataSet:         metadataSet,
		FileDownloads:       *NewFileDownloads(),
		key:                 key,
		reputationTable:     reptable,
		trustedKeys:         trustedKeys,
		keyRing:             awot.NewKeyRing(parameters.Identifier, key.PublicKey, trustedKeys),
	}
	gossiper.keyRing.StartWithReputation(time.Duration(5)*time.Second, &reptable)
	return &gossiper
}

// Start the Gossiper
func (g *Gossiper) Start() {

	var wg sync.WaitGroup
	wg.Add(8)

	// Standard output writer Thread
	go func() {
		defer wg.Done()
		fmtwriter(g.standardOutputQueue)
	}()
	// Client Listener Thread
	go func() {
		defer wg.Done()
		listener(g.Parameters.UIConn, g, handleClientMessage)
	}()
	// Client Writer Thread
	go func() {
		defer wg.Done()
		clientwriter(g.Parameters.UIConn, g.clientOutputQueue)
	}()
	// Gossiper Listener Thread
	go func() {
		defer wg.Done()
		listener(g.Parameters.GossipConn, g, handleGossiperMessage)
	}()
	// Gossiper Writer Thread
	go func() {
		defer wg.Done()
		writer(g, g.Parameters.GossipConn, g.gossipOutputQueue)
	}()

	// Anti Entropy Thread
	go func() {
		defer wg.Done()
		antiEntropy(g, g.Parameters.Etimer)
	}()

	// Route Rumor Sender Thread
	go func() {
		defer wg.Done()
		routerumor(g, g.Parameters.Rtimer)
	}()

	// Reputation Update Requests Thread
	go func() {
		defer wg.Done()
		repUpdateRequests(g, g.Parameters.Reptimer)
	}()

	// Reputation Logs Thread
	go func() {
		defer wg.Done()
		repLogs(g)
	}()

	fmt.Println("INITIALIZATION DONE")

	// Broadcast a route rumor message
	broadcastNewRoute(g)

	time.Sleep(1000 * time.Millisecond)

	// Send signatures
	g.SendSignatures()

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

	// Initialize A's contrib-based reputation if necessary
	g.reputationTable.InitContribRepForPeer(addrToString(A.Address))

	// demultiplex packets
	if pkt.Rumor != nil {
		// process rumor
		g.processRumor(pkt.Rumor, remoteaddr)
	}
	if pkt.Status != nil {
		// process status
		go g.processStatus(pkt.Status, remoteaddr)
	}
	if pkt.Private != nil {
		// process private message
		go g.processPrivateMessage(pkt.Private, remoteaddr)
	}
	if pkt.DataRequest != nil {
		// process data request
		go g.processDataRequest(pkt.DataRequest, remoteaddr)
	}
	if pkt.DataReply != nil {
		// process data reply
		go g.processDataReply(pkt.DataReply, remoteaddr)
	}
	if pkt.RepContribUpdateReq {
		// process contrib-based reputation update request
		go g.processContribRepUpdateReq(&A)
	}
	if pkt.RepUpdate != nil {
		// process contrib-based reputation update
		go g.processContribRepUpdate(pkt.RepUpdate, &A)
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
