package main

/*
    Imports
*/

import (

  "sync"
  "time"

  "github.com/No-Trust/peerster/common"
  "github.com/No-Trust/peerster/rep"

)

/*
    Functions
*/

func repUpdateRequests(g *Gossiper, reptimer uint, wg sync.WaitGroup) {

  defer wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(reptimer))
	defer ticker.Stop()

	for _ = range ticker.C {

    log := "Sending reputation update requests to most reputable peers..."

    g.standardOutputQueue <- &log

    highestRepTable := g.reputationTable.MostReputablePeers(rep.REP_REQ_PEER_COUNT)

    highestRepTable.ForEachSigRep(func(peer string, _ float32) {

      nextHop := g.routingTable.Get(peer)

      if nextHop != "" {

        g.gossipOutputQueue <- &Packet {
          GossipPacket : GossipPacket {
            Private : &PrivateMessage {
              RepSigUpdateReq : true,
            },
          },
          Destination: stringToUDPAddr(nextHop),
        }

      }

    })

    highestRepTable.ForEachContribRep(func(peer string, _ float32) {

      g.gossipOutputQueue <- &Packet {
        GossipPacket : GossipPacket {
          RepContribUpdateReq : true,
        },
        Destination  : stringToUDPAddr(peer),
      }

    })

	}

}

func repLogs(g *Gossiper, wg sync.WaitGroup) {

  defer wg.Done()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

  for _ = range ticker.C {
    g.reputationTable.Log()
  }

}

func (g *Gossiper) processContribRepUpdateReq(sender *common.Peer) {

  log := "SENDING CONTRIB-REP UPDATE TO " + addrToString(sender.Address)
  g.standardOutputQueue <- &log

  g.gossipOutputQueue <- &Packet {
		GossipPacket : GossipPacket {
			RepUpdate : g.reputationTable.GetContribUpdate(),
		},
		Destination  : sender.Address,
	}

}

func (g *Gossiper) processContribRepUpdate(update *rep.RepUpdate, sender *common.Peer) {

  log := "RECEIVED CONTRIB-REP UPDATE FROM " + addrToString(sender.Address) + "\nPRINTING OLD REPS AND NEW REPS"
  g.standardOutputQueue <- &log
  g.reputationTable.Log()

  g.reputationTable.UpdateReputations(update, addrToString(sender.Address))

  g.reputationTable.Log()

}
