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

    highestRepTable.ForEachSigRep(func(peer *common.Peer, _ float32) {

      g.gossipOutputQueue <- &Packet {
        GossipPacket : GossipPacket {
          RepUpdateReq : &rep.RepUpdateRequest {
            SigUpdateReq : true,
          },
        },
        Destination  : peer.Address,
      }

    })

    highestRepTable.ForEachContribRep(func(peer *common.Peer, _ float32) {

      g.gossipOutputQueue <- &Packet {
        GossipPacket : GossipPacket {
          RepUpdateReq : &rep.RepUpdateRequest {
            ContribUpdateReq : true,
          },
        },
        Destination  : peer.Address,
      }

    })

	}

}

func (g *Gossiper) processRepUpdateReq(request *rep.RepUpdateRequest, sender *common.Peer) {

  var repUpdate *rep.RepUpdate

  if request.SigUpdateReq {
    repUpdate = g.reputationTable.GetSigUpdate()
  } else if request.ContribUpdateReq {
    repUpdate = g.reputationTable.GetContribUpdate()
  } else {

    err := "ERROR: Invalid reputation update request."

    g.standardOutputQueue <- &err

    return

  }

  g.gossipOutputQueue <- &Packet {
		GossipPacket : GossipPacket {
			RepUpdate : repUpdate,
		},
		Destination  : sender.Address,
	}

}
