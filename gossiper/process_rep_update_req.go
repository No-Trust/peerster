package main

/*
   Imports
*/

import (
	"time"

	"github.com/No-Trust/peerster/common"
	"github.com/No-Trust/peerster/rep"
)

/*
   Functions
*/

func repUpdateRequests(g *Gossiper, reptimer uint) {

	ticker := time.NewTicker(time.Second * time.Duration(reptimer))
	defer ticker.Stop()

	for range ticker.C {

		common.Log("Sending reputation update requests to most reputable peers...",
			common.LOG_MODE_FULL)

		highestRepTable := g.reputationTable.MostReputablePeers(rep.REP_REQ_PEER_COUNT)

		highestRepTable.ForEachSigRep(func(peer string, _ float32) {

			nextHop := g.routingTable.Get(peer)

			if nextHop != "" {

				g.gossipOutputQueue <- &Packet{
					GossipPacket: GossipPacket{
						Private: &PrivateMessage{
							RepSigUpdateReq: true,
						},
					},
					Destination: stringToUDPAddr(nextHop),
				}

			}

		})

		highestRepTable.ForEachContribRep(func(peer string, _ float32) {

			g.gossipOutputQueue <- &Packet{
				GossipPacket: GossipPacket{
					RepContribUpdateReq: true,
				},
				Destination: stringToUDPAddr(peer),
			}

		})

	}

}

func repLogs(g *Gossiper) {

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		g.reputationTable.Log()
	}

}

func (g *Gossiper) processContribRepUpdateReq(sender *common.Peer) {

	common.Log("SENDING CONTRIB-REP UPDATE TO "+addrToString(sender.Address),
		common.LOG_MODE_FULL)

	g.gossipOutputQueue <- &Packet{
		GossipPacket: GossipPacket{
			RepUpdate: g.reputationTable.GetContribUpdate(),
		},
		Destination: sender.Address,
	}

}

func (g *Gossiper) processContribRepUpdate(update *rep.RepUpdate, sender *common.Peer) {

	common.Log("RECEIVED CONTRIB-REP UPDATE FROM "+addrToString(sender.Address),
		common.LOG_MODE_FULL)

	g.reputationTable.UpdateReputations(update, addrToString(sender.Address))

}
