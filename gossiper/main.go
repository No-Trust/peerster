// Main function, deals with user paramaters and launch the gossiper peer
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"net"
	"strconv"
	"strings"
)

const HOP_LIMIT = 10
const CHANNEL_SIZE = 100
const CHUNK_SIZE = 8000
const FILES_DIR = "../_Downloads/"
const CHUNKS_DIR = "../_Downloads/.Chunks/"
const HASH_LENGTH = 256
const KEY_FILE_NAME = "../private.key"

// Main
func main() {

	UIPort := flag.Uint("UIPort", 10000, "port for the UI client")
	gossipIPPort := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")

	rtimer := flag.Uint("rtimer", 60, "timer duration for the sending of route rumors")
	etimer := flag.Uint("etimer", 2, "timer duration for the sending of anti entropy status")
	noforward := flag.Bool("noforward", false, "for testing : forwarding of route rumors only")
	nat_traversal := flag.Bool("traversal", false, "nat travarsal option")

	flag.Parse()

	fmt.Println("given peers :", *peers, "\n")

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
		Identifier:      identifier,
		Name:            *name,
		Etimer:          *etimer,
		Rtimer:          *rtimer,
		Hoplimit:        HOP_LIMIT,
		NoForward:       *noforward,
		NatTraversal:    *nat_traversal,
		GossipAddr:      *gossipAddr,
		GossipConn:      *gossipConn,
		UIAddr:          *UIAddr,
		UIConn:          *UIConn,
		ChannelSize:     CHANNEL_SIZE,
		ChunkSize:       CHUNK_SIZE,
		FilesDirectory:  FILES_DIR,
		ChunksDirectory: CHUNKS_DIR,
		HashLength:      HASH_LENGTH,
		KeyFileName:     KEY_FILE_NAME,
	}

	var g *Gossiper = NewGossiper(parameters, peerAddrs)

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
		var ip net.IP = net.ParseIP(ipS)
		if ip == nil {
			common.CheckRead(errors.New("ip address must be correct"))
		}
		var newPeer net.UDPAddr = net.UDPAddr{ip, port, ""}
		peers = append(peers, newPeer)
	}
	return peers
}
