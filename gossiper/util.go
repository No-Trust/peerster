// Util functions
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/No-Trust/peerster/common"
	"github.com/dedis/protobuf"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func flipCoin() bool {
	return (rand.Int() % 2) == 0
}

func addrToString(addr net.UDPAddr) string {
	return addr.IP.String() + ":" + strconv.Itoa(addr.Port)
}

func stringToUDPAddr(ipport string) net.UDPAddr {
	ipS, portS, err := net.SplitHostPort(ipport)
	common.CheckRead(err)
	port, err := strconv.Atoi(portS)
	common.CheckRead(err)
	var ip = net.ParseIP(ipS)
	if ip == nil {
		common.CheckRead(errors.New("ip address must be correct"))
	}
	var addr = net.UDPAddr{
		IP:   ip,
		Port: port,
		Zone: "",
	}
	return addr
}

func UDPAddrToString(addr net.UDPAddr) string {
	ipport := addr.IP.String() + ":" + strconv.Itoa(addr.Port)
	return ipport
}

// identifier string for an ack
func AckString(addr net.UDPAddr, origin string, nextID uint32) string {
	return fmt.Sprintf("%s | %s | %d", UDPAddrToString(addr), origin, nextID)
}

// Writer for gossiper packets : send every packet coming for a channel to the destination
func writer(g *Gossiper, udpConn net.UDPConn, queue chan *Packet) {
	// writing loop
	// write every message on queue
	for pkt := range queue {
		destination := pkt.Destination
		gossipPacket := pkt.GossipPacket

		buf, err := protobuf.Encode(&gossipPacket)
		common.CheckRead(err)
		_, err = udpConn.WriteToUDP(buf, &destination)
		common.CheckRead(err)
	}
}

// Writer for client packets : send every packet coming for a channel to the destination
func clientwriter(udpConn net.UDPConn, queue chan *common.Packet) {
	// writing loop
	// write every message on queue
	for pkt := range queue {
		serverpkt := pkt.ClientPacket
		destination := pkt.Destination

		buf, err := protobuf.Encode(&serverpkt)
		common.CheckRead(err)
		_, err = udpConn.WriteToUDP(buf, &destination)
		common.CheckRead(err)
	}
}

// Listening Loop, calls handler when there is a new packet
func listener(udpConn net.UDPConn, g *Gossiper, handler func([]byte, *net.UDPAddr, *Gossiper)) {
	defer udpConn.Close()

	buf := make([]byte, 65535) // receiving byte array

	// Listening loop
	for {
		n, remoteaddr, err := udpConn.ReadFromUDP(buf)
		common.CheckRead(err)
		handler(buf[:n], remoteaddr, g)
	}
}

// Write bytes to disk, at location directory with name filename
func writeToDisk(data []byte, directory, filename string) {
	os.MkdirAll(directory, os.ModePerm)
	path := directory + filename
	f, err := os.Create(path)
	common.CheckError(err)
	defer f.Close()
	_, err = f.Write(data)
	f.Sync()
}

// Split a byte slice into chunks of specified size
func splitInChunks(data []byte, chunksize uint) *[][]byte {
	length := len(data)
	remaining := uint(length)

	chunks := make([][]byte, 0)

	var i uint
	for i = 0; remaining > 0; i += chunksize {
		j := chunksize
		if j > remaining {
			j = remaining
		}

		chunks = append(chunks, data[i:i+j])
		remaining -= j
	}

	return &chunks
}

// Reassemble the chunks into one slice
func reassembleChunks(chunks *[][]byte) *[]byte {
	r := make([]byte, 0)
	for i := 0; i < len(*chunks); i++ {
		r = append(r, (*chunks)[i]...)
	}

	return &r
}

// hash each chunk
func hashChunks(chunks *[][]byte) [][]byte {
	hashes := make([][]byte, 0)

	for i := 0; i < len(*chunks); i++ {
		h := sha256.New()
		h.Write((*chunks)[i])
		hashes = append(hashes, h.Sum(nil))
	}

	return hashes
}

func GetChunkFilename(chunk []byte) string {
	h := sha256.New()
	h.Write(chunk)
	hash := h.Sum(nil)
	hashstr := hex.EncodeToString(hash)
	if len(hashstr) > 128 {
		hashstr = hashstr[:128]
	}
	filename := hashstr + ".chunk"
	return filename
}

func GetChunkFilenameFromHash(hash []byte, hashlen uint) string {
	offset := hashlen / 8

	hashstr := hex.EncodeToString(hash[:offset])
	if len(hashstr) > 128 {
		hashstr = hashstr[:128]
	}
	filename := hashstr + ".chunk"
	return filename
}

func writeChunksToDisk(chunks [][]byte, directory, filename string) {
	// format : hash[0:128].chunk

	path, err := filepath.Abs("")
	common.CheckError(err)
	for _, chunk := range chunks {
		// storing chunk i
		filename := GetChunkFilename(chunk)
		chunkDir := path + string(os.PathSeparator) + directory
		writeToDisk(chunk, chunkDir, filename)
	}
}
