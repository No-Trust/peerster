package main

import (
  "net"
  "github.com/dedis/protobuf"
  "github.com/No-Trust/peerster/common"
)

// Listening Loop, calls handler when there is a new packet
func listener(udpConn net.UDPConn, handler func([]byte, *net.UDPAddr)) {
	defer udpConn.Close()

	buf := make([]byte, 65535) // receiving byte array

	// Listening loop
	for {
  	n, remoteaddr, err := udpConn.ReadFromUDP(buf)
		common.CheckError(err)
		//go handler(buf[:n], remoteaddr, g)
    handler(buf[:n], remoteaddr)
	}
}

func writer(udpConn net.UDPConn, queue chan *common.ClientPacket, destination *net.UDPAddr) {
	// writing loop
	// write every message on queue
  defer udpConn.Close()
	for pkt := range queue {

		buf, err := protobuf.Encode(pkt)
		common.CheckError(err)
		_, err = udpConn.WriteToUDP(buf, destination)
    common.CheckError(err)
	}
}
