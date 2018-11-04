// Main function, deals with user parameters and launch the gossiper peer
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/No-Trust/peerster/common"
	"github.com/No-Trust/peerster/rep"
)

const HOP_LIMIT = 10
const CHANNEL_SIZE = 100
const CHUNK_SIZE = 8000
const FILES_DIR = "../_Downloads/"
const CHUNKS_DIR = "../_Downloads/.Chunks/"
const HASH_LENGTH = 256
const KEY_DIRECTORY = "../"

// Main
func main() {

	UIPort := flag.Uint("UIPort", 10000, "port for the UI client")
	gossipIPPort := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")

	rtimer := flag.Uint("rtimer", 60, "timer duration for the sending of route rumors")
	etimer := flag.Uint("etimer", 2, "timer duration for the sending of anti entropy status")
	reptimer := flag.Uint("reptimer", rep.DEFAULT_REP_REQ_TIMER,
		"timer duration for reputation update requests")
	noforward := flag.Bool("noforward", false, "for testing : forwarding of route rumors only")
	natTraversal := flag.Bool("traversal", false, "nat travarsal option")
	keysdir := flag.String("keys", ".", "directory for boostrap public keys")
	confidenceThreshold := flag.Float64("cthresh", 0.20, "confidence threshold for collected public keys")

	// Program execution log mode
	logMode := flag.String("logs", common.LOG_MODE_REACTIVE, "execution log mode")

	flag.Parse()

	common.InitLogger(*logMode)

	fmt.Println("given peers :", *peers)

	sipport := strings.Split(*gossipIPPort, ":")
	if len(sipport) < 2 {
		common.CheckRead(errors.New("gossipPort must be of the form ip:port"))
	}

	gossipIP := sipport[0]
	gossipPort := sipport[1]

	if *name == "" {
		*name = "peerster@" + gossipIP + ":" + gossipPort
	}
	var peerAddrs []net.UDPAddr
	if *peers != "" {
		peerAddrs = parsePeers(strings.Split(*peers, ","))
	}

	UIAddress := fmt.Sprintf("%s:%d", "127.0.0.1", *UIPort)

	fmt.Printf("Gossiper '%s' started \n", *name)
	fmt.Printf("Listening for client on : %s\n", UIAddress)
	fmt.Printf("Listening for other peerster on : %s \n", *gossipIPPort)
	fmt.Println("With peers :", peerAddrs)

	if *noforward {
		fmt.Println("Not forwarding text rumor")
		fmt.Println("Not forwarding private message")
	}

	gossipAddress := fmt.Sprintf("%s:%s", gossipIP, gossipPort)
	identifier := *name // TODO !!

	// Opening gossip socket
	gossipAddr, err := net.ResolveUDPAddr("udp4", gossipAddress)
	common.CheckError(err)
	gossipConn, err := net.ListenUDP("udp4", gossipAddr)
	common.CheckError(err)

	// Opening client socket
	UIAddr, err := net.ResolveUDPAddr("udp4", UIAddress)
	common.CheckError(err)
	UIConn, err := net.ListenUDP("udp4", UIAddr)
	common.CheckError(err)

	parameters := Parameters{
		Identifier:             identifier,
		Name:                   *name,
		Etimer:                 *etimer,
		Rtimer:                 *rtimer,
		Reptimer:               *reptimer,
		Hoplimit:               HOP_LIMIT,
		NoForward:              *noforward,
		NatTraversal:           *natTraversal,
		GossipAddr:             *gossipAddr,
		GossipConn:             *gossipConn,
		UIAddr:                 *UIAddr,
		UIConn:                 *UIConn,
		ChannelSize:            CHANNEL_SIZE,
		ChunkSize:              CHUNK_SIZE,
		FilesDirectory:         FILES_DIR,
		ChunksDirectory:        CHUNKS_DIR,
		HashLength:             HASH_LENGTH,
		KeyFileName:            KEY_DIRECTORY + "private.key",
		PubKeyFileName:         KEY_DIRECTORY + identifier + ".pub",
		TrustedKeysDirectory:   *keysdir,
		KeyConfidenceThreshold: float32(*confidenceThreshold),
	}

	var g = NewGossiper(parameters, peerAddrs)

	// start peerster
	g.Start()
}

// from array of string ip:port return the array of UDPAddr
func parsePeers(args []string) []net.UDPAddr {
	// parses slice of strings of the form ip:port into slice of UDPAddr
	peers := make([]net.UDPAddr, 0)
	for _, v := range args {
		ipS, portS, err := net.SplitHostPort(v)
		common.CheckRead(err)
		port, err := strconv.Atoi(portS)
		common.CheckRead(err)
		var ip = net.ParseIP(ipS)
		if ip == nil {
			common.CheckRead(errors.New("ip address must be correct"))
		}
		var newPeer = net.UDPAddr{IP: ip, Port: port, Zone: ""}
		peers = append(peers, newPeer)
	}
	return peers
}
