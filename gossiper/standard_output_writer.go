// Writer for the standard output
package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

func fmtwriter(queue chan *string, wg sync.WaitGroup) {
	// writer for the standard output
	// write every string on queue
	defer wg.Done()

	for str := range queue {
		fmt.Println(*str)
	}
}

// Strings for messages

func DirectRouteString(origin string, remoteaddr *net.UDPAddr) *string {
	str := fmt.Sprintf("DIRECT-ROUTE FOR %s: %s", origin, addrToString(*remoteaddr))
	return &str
}

func (msg *SimpleMessage) SimpleMessageString() *string {
	str := fmt.Sprintf("CLIENT %s %s", msg.Text, msg.SenderName)
	return &str
}

func (msg *RumorMessage) RumorString(source *net.UDPAddr) *string {
	str := ""
	if msg.Text == "" {
		// route rumor
		//str := fmt.Sprintf("ROUTE RUMOR origin %s from %s:%s ID %d", msg.Origin, source.IP.String(), strconv.Itoa(source.Port), msg.ID)
		str += fmt.Sprintf("DSDV %s:%s:%d \n", msg.Origin, source.IP.String(), source.Port)
	}
	// rumor message
	str += fmt.Sprintf("RUMOR origin %s from %s:%s ID %d contents %s", msg.Origin, source.IP.String(), strconv.Itoa(source.Port), msg.ID, msg.Text)
	return &str
}

func (msg *RumorMessage) MongeringString(dest *net.UDPAddr) *string {
	rumorType := "TEXT"
	if msg.Text == "" {
		rumorType = "ROUTE"
	}
	str := fmt.Sprintf("MONGERING %s with %s:%s", rumorType, dest.IP.String(), strconv.Itoa(dest.Port))
	return &str
}

func (msg *StatusPacket) StatusString(source *net.UDPAddr) *string {
	origins := ""
	for _, peerstatus := range msg.Want {
		origins += fmt.Sprintf(" origin %s nextID %d", peerstatus.Identifier, peerstatus.NextID)
	}
	str := fmt.Sprintf("STATUS from %s:%s%s", source.IP.String(), strconv.Itoa(source.Port), origins)
	return &str
}

func CoinFlipString(dest *net.UDPAddr) *string {
	str := fmt.Sprintf("FLIPPED COIN sending rumor to %s:%s", dest.IP.String(), strconv.Itoa(dest.Port))
	return &str
}

func SyncString(peer *net.UDPAddr) *string {
	str := fmt.Sprintf("IN SYNC WITH %s:%s", peer.IP.String(), strconv.Itoa(peer.Port))
	return &str
}

func (msg *StatusPacket) AntiEntropyString(peer *net.UDPAddr) *string {
	origins := ""
	for _, peerstatus := range msg.Want {
		origins += fmt.Sprintf(" origin %s nextID %d", peerstatus.Identifier, peerstatus.NextID)
	}
	str := fmt.Sprintf("ANTI ENTROPY STATUS to %s:%s %s", peer.IP.String(), strconv.Itoa(peer.Port), origins)
	return &str
}

func (pm *PrivateMessage) PrivateMessageString(source *net.UDPAddr) *string {
	//str := fmt.Sprintf("PRIVATE MESSAGE origin %s from %s:%s contents %s", pm.Origin, source.IP.String(), strconv.Itoa(source.Port), pm.Text)
	str := fmt.Sprintf("PRIVATE: %s:%d:%s", pm.Origin, pm.HopLimit, pm.Text)
	return &str
}

func (req *DataRequest) DataRequestString(source *net.UDPAddr) *string {
  str := fmt.Sprintf("DATA REQUEST: %s:%d:%s:%s", req.Origin, req.HopLimit, req.FileName, string(req.HashValue))
  return &str
}
