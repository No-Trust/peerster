// Messages
package main

import (
	"github.com/No-Trust/peerster/common"
	"net"
	"sync"
	"fmt"
	"github.com/No-Trust/peerster/awot"
)

/*
 * Application Packets used for peer communications
 */

type Message struct {
	Text string
}

type SimpleMessage struct {
	SenderName string
	Text       string
}

type PeerMessage struct {
	ID   uint32
	Text string
}

type RumorMessage struct {
	Origin   string
	ID       uint32
	Text     string
	LastIP   *net.IP
	LastPort *int
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want  []PeerStatus
	mutex *sync.Mutex
}

/***** Private Message *****/

type PrivateMessage struct {
	Origin   string
	ID       uint32
	Text     string
	Dest     string
	HopLimit uint32
}

/***** Data Request & Reply *****/

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	FileName    string
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	FileName    string
	HashValue   []byte
	Data        []byte
}

type GossipPacket struct {
	Rumor       *RumorMessage
	Status      *StatusPacket
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
	KeyExchange *awot.KeyExchangeMessage
}

type Packet struct {
	GossipPacket GossipPacket
	Destination  net.UDPAddr
}

/***** Rumor Message *****/

func (rumor *RumorMessage) isRoute() bool {
	return rumor.Text == ""
}

func (rumor *RumorMessage) isChat() bool {
	return !rumor.isRoute()
}

/***** Peer Status *****/

func (sp *StatusPacket) Length() int {
	sp.mutex.Lock()
	l := len(sp.Want)
	sp.mutex.Unlock()
	return l
}

func (sp *StatusPacket) GetIndex(i int) PeerStatus {
	// unsafe
	sp.mutex.Lock()
	ps := sp.Want[i]
	sp.mutex.Unlock()
	return ps
}

func (sp *StatusPacket) Copy() *StatusPacket {
	// return a copy of the statuspacket

	sp.mutex.Lock()
	Wantcopy := make([]PeerStatus, len(sp.Want))
	copy(Wantcopy, sp.Want)
	sp.mutex.Unlock()
	spc := StatusPacket{
		Want:  Wantcopy,
		mutex: &sync.Mutex{},
	}
	return &spc
}

func (sp *StatusPacket) Get(identifier string) uint32 {
	sp.mutex.Lock()
	for _, peerstatus := range sp.Want {
		if peerstatus.Identifier == identifier {
			NextID := peerstatus.NextID
			sp.mutex.Unlock()
			return NextID
		}
	}
	// havent found it : add it
	sp.Want = append(sp.Want, PeerStatus{
		Identifier: identifier,
		NextID:     1,
	})
	sp.mutex.Unlock()
	return 1
}

// Add 1 to entry
func (sp *StatusPacket) Update(identifier string) {
	sp.mutex.Lock()
	for index, peerstatus := range sp.Want {
		if peerstatus.Identifier == identifier {
			sp.Want[index].NextID += 1
			sp.mutex.Unlock()
			return
		}
	}
	// theres no identifier in the slice
	// so create it
	sp.Want = append(sp.Want, PeerStatus{
		Identifier: identifier,
		NextID:     2,
	})
	sp.mutex.Unlock()
}

func (sp *StatusPacket) Add(identifier string) {
	if sp.Get(identifier) <= 0 {
		sp.mutex.Lock()
		sp.Want = append(sp.Want, PeerStatus{
			Identifier: identifier,
			NextID:     1,
		})
		sp.mutex.Unlock()
	}
}

func NewStatusPacket(peers []common.Peer, identifier string) *StatusPacket {
	var Want []PeerStatus

	// // Adding peers clocks
	// for _, peer := range peers {
	// 	Want = append(Want, PeerStatus{
	// 		Identifier: peer.Identifier,
	// 		NextID:     1,
	// 	})
	// }
	// Adding its own clock
	Want = append(Want, PeerStatus{
		Identifier: identifier,
		NextID:     1,
	})
	sp := StatusPacket{Want, &sync.Mutex{}}
	return &sp
}

func (sp *StatusPacket) String() string {
	sp.mutex.Lock()
	str := ""
	for _, status := range sp.Want {
		str += fmt.Sprintf("id: %s, next msg: %d \n", status.Identifier, status.NextID)
	}
	sp.mutex.Unlock()
	return str
}
